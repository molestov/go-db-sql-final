package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testDriverName   = "sqlite"
	testDatabaseName = "tracker.db"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

type TestSuite struct {
	suite.Suite
	db *sql.DB
}

func (suite *TestSuite) SetupTest() {
	db, err := sql.Open(testDriverName, testDatabaseName)
	require.Empty(suite.T(), err)
	suite.db = db
}

func (suite *TestSuite) TearDownTest() {
	err := suite.db.Close()
	if err != nil {
		return
	}
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

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
func (suite *TestSuite) TestAddGetDelete() {
	// prepare
	store := NewParcelStore(suite.db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	require.Empty(suite.T(), err)
	require.NotEmpty(suite.T(), number)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	storedParcel, err := store.Get(number)
	require.Empty(suite.T(), err)
	assert.Equal(suite.T(), number, storedParcel.Number)
	assert.Equal(suite.T(), parcel.Client, storedParcel.Client)
	assert.Equal(suite.T(), parcel.Address, storedParcel.Address)
	assert.Equal(suite.T(), parcel.Status, storedParcel.Status)
	assert.Equal(suite.T(), parcel.CreatedAt, storedParcel.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(number)
	require.Empty(suite.T(), err)

	_, err = store.Get(number)
	require.NotEmpty(suite.T(), err)
}

// TestSetAddress проверяет обновление адреса
func (suite *TestSuite) TestSetAddress() {
	// prepare
	store := NewParcelStore(suite.db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	require.Empty(suite.T(), err)
	require.NotEmpty(suite.T(), number)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	require.Empty(suite.T(), err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	storedParcel, err := store.Get(number)
	require.Empty(suite.T(), err)
	assert.Equal(suite.T(), number, storedParcel.Number)
	assert.Equal(suite.T(), parcel.Client, storedParcel.Client)
	assert.Equal(suite.T(), newAddress, storedParcel.Address)
	assert.Equal(suite.T(), parcel.Status, storedParcel.Status)
	assert.Equal(suite.T(), parcel.CreatedAt, storedParcel.CreatedAt)
}

// TestSetStatus проверяет обновление статуса
func (suite *TestSuite) TestSetStatus() {
	// prepare
	store := NewParcelStore(suite.db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	require.Empty(suite.T(), err)
	require.NotEmpty(suite.T(), number)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusSent
	err = store.SetStatus(number, newStatus)
	require.Empty(suite.T(), err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	storedParcel, err := store.Get(number)
	require.Empty(suite.T(), err)
	assert.Equal(suite.T(), number, storedParcel.Number)
	assert.Equal(suite.T(), parcel.Client, storedParcel.Client)
	assert.Equal(suite.T(), parcel.Address, storedParcel.Address)
	assert.Equal(suite.T(), newStatus, storedParcel.Status)
	assert.Equal(suite.T(), parcel.CreatedAt, storedParcel.CreatedAt)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func (suite *TestSuite) TestGetByClient() {
	// prepare
	store := NewParcelStore(suite.db)

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
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		number, err := store.Add(parcels[i])
		require.Empty(suite.T(), err)
		require.NotEmpty(suite.T(), number)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = number

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[number] = parcels[i]
	}

	// get by client
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	storedParcels, err := store.GetByClient(client)
	// убедитесь в отсутствии ошибки
	require.Empty(suite.T(), err)
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	assert.Equal(suite.T(), len(parcels), len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		addedParcel, ok := parcelMap[parcel.Number]
		require.True(suite.T(), ok)
		assert.Equal(suite.T(), addedParcel.Number, parcel.Number)
		assert.Equal(suite.T(), addedParcel.Client, parcel.Client)
		assert.Equal(suite.T(), addedParcel.Address, parcel.Address)
		assert.Equal(suite.T(), addedParcel.Status, parcel.Status)
		assert.Equal(suite.T(), addedParcel.CreatedAt, parcel.CreatedAt)

	}
}
