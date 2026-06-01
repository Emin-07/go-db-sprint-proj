package main

import (
	"context"
	"database/sql"
	"math/rand"
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
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	ctx := context.Background()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(ctx, parcel)
	assert.Nil(t, err)
	require.NotEqual(t, id, 0)
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	selectedParcel, err := store.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, selectedParcel.Number, id)
	assert.Equal(t, selectedParcel.Client, parcel.Client)
	assert.Equal(t, selectedParcel.Status, parcel.Status)
	assert.Equal(t, selectedParcel.Address, parcel.Address)
	assert.Equal(t, selectedParcel.CreatedAt, parcel.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(ctx, selectedParcel.Number)
	assert.Nil(t, err)

	parcelAfterDelete, err := store.Get(ctx, id)
	assert.NotNil(t, err)
	assert.Empty(t, parcelAfterDelete.Number)
	assert.Empty(t, parcelAfterDelete.Client)
	assert.Empty(t, parcelAfterDelete.Status)
	assert.Empty(t, parcelAfterDelete.Address)
	assert.Empty(t, parcelAfterDelete.CreatedAt)

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	ctx := context.Background()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(ctx, parcel)
	require.Nil(t, err)
	require.NotEqual(t, id, 0)
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(ctx, id, newAddress)
	assert.Nil(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	changedParcel, err := store.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, changedParcel.Address, newAddress)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	ctx := context.Background()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(ctx, parcel)
	require.Nil(t, err)
	require.NotEqual(t, id, 0)
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(ctx, id, ParcelStatusDelivered)
	assert.Nil(t, err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	changedParcel, err := store.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, changedParcel.Status, ParcelStatusDelivered)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	ctx := context.Background()

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(ctx, parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		assert.Nil(t, err)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(ctx, client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	assert.Nil(t, err)
	assert.Equal(t, len(storedParcels), len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		testParcel, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		assert.Equal(t, testParcel.Number, parcel.Number)
		assert.Equal(t, testParcel.Client, parcel.Client)
		assert.Equal(t, testParcel.Status, parcel.Status)
		assert.Equal(t, testParcel.Address, parcel.Address)
		assert.Equal(t, testParcel.CreatedAt, parcel.CreatedAt)
	}
}
