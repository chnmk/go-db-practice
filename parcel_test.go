package main

import (
	"database/sql"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// Настройка подключения к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// Добавляем новую посылку в БД
	id, err := store.Add(parcel)

	// Проверка на отсутствие ошибки и наличие идентификатора
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// get
	// Получаем только что добавленную посылку, убеждаемся в отсутствии ошибки
	p, err := store.Get(id)
	require.NoError(t, err)

	// Проверяем, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	assert.Equal(t, parcel.Client, p.Client)
	assert.Equal(t, parcel.Status, p.Status)
	assert.Equal(t, parcel.Address, p.Address)
	assert.Equal(t, parcel.CreatedAt, p.CreatedAt)

	// delete
	// Удаляем добавленную посылку, убеждаемся в отсутствии ошибки
	err = store.Delete(id)
	require.NoError(t, err)

	// Проверяем, что посылку больше нельзя получить из БД
	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// Настройка подключения к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// Добавляем новую посылку в БД
	id, err := store.Add(parcel)

	// Проверка на отсутствие ошибки и наличие идентификатора
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// set address
	// Обновляем адрес
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)

	// Проверка на отсутствие ошибки
	require.NoError(t, err)

	// check
	// Получаем добавленную посылку
	p, err := store.Get(id)

	// Смотрим, обновился ли адрес
	require.NoError(t, err)
	assert.Equal(t, newAddress, p.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// Настройка подключения к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// Добавляем новую посылку в БД
	id, err := store.Add(parcel)

	// Проверка на отсутствие ошибки и наличие идентификатора
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// set status
	// Обновляем статус
	err = store.SetStatus(id, ParcelStatusDelivered)

	// Проверка на отсутствие ошибки
	require.NoError(t, err)

	// check
	// Получаем добавленную посылку
	p, err := store.Get(id)

	// Смотрим, обновился ли статус
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusDelivered, p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// Настройка подключения к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)

	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// Задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		// Добавляем новую посылку в БД
		id, err := store.Add(parcels[i])

		// Проверка на отсутствие ошибки и наличие идентификатора
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// Обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// Сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// Получаем список посылок по идентификатору клиента, сохранённого в переменной client
	storedParcels, err := store.GetByClient(client)

	// Проверяем, отсутствует ли ошибка
	require.NoError(t, err)
	// Смотрим, что количество полученных посылок совпадает с количеством добавленных
	assert.Equal(t, len(storedParcels), len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// Пробуем получить посылку из мапы по текущему ключу
		parcelFromMap, exists := parcelMap[parcel.Number]

		// Смотрим, что посылка из storedParcels есть в parcelMap
		require.True(t, exists)

		// Проверяем, что значения полей заполнены верно
		assert.True(t, reflect.DeepEqual(parcelFromMap, parcel))
	}
}
