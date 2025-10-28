package discriminator

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
)

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestStrictServerGetPets(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	req := httptest.NewRequest(http.MethodGet, "/pets", nil)
	w := httptest.NewRecorder()

	handler.GetPets(w, req)

	expectedJSON := `[
		{"id":"1","petType":"cat","name":"Whiskers","meow":true},
		{"id":"2","petType":"dog","name":"Buddy","bark":true},
		{"id":"3","petType":"cat","name":"Mittens","meow":false}
	]`
	assert.JSONEq(t, expectedJSON, w.Body.String())
}

func TestStrictServerGetPet(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	tests := []struct {
		name         string
		petId        string
		expectedJSON string
	}{
		{
			name:         "Cat",
			petId:        "cat-1",
			expectedJSON: `{"id":"cat-1","petType":"cat","name":"Whiskers","meow":true}`,
		},
		{
			name:         "Dog",
			petId:        "dog-1",
			expectedJSON: `{"id":"dog-1","petType":"dog","name":"Buddy","bark":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/pets/"+tt.petId, nil)
			w := httptest.NewRecorder()

			handler.GetPet(w, req, tt.petId)

			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

func TestStrictServerCreatePet(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	tests := []struct {
		name         string
		petJSON      string
		expectedJSON string
	}{
		{
			name:         "Cat",
			petJSON:      `{"petType": "cat", "name": "Fluffy", "meow": true}`,
			expectedJSON: `{"id": "new-cat", "petType": "cat", "name": "Fluffy", "meow": true}`,
		},
		{
			name:         "Dog",
			petJSON:      `{"petType": "dog", "name": "Rex", "bark": true}`,
			expectedJSON: `{"id": "new-dog", "petType": "dog", "name": "Rex", "bark": true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/pets", bytes.NewBufferString(tt.petJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreatePet(w, req)

			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

func TestStrictServerGetActivities(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	req := httptest.NewRequest(http.MethodGet, "/pets/pet-1/activities", nil)
	w := httptest.NewRecorder()

	handler.GetActivities(w, req, "pet-1")

	expectedJSON := `[
		{"id":"1","activityType":"feeding","duration":30,"foodType":"kibble"},
		{"id":"2","activityType":"walking","duration":60,"distance":2.5}
	]`
	assert.JSONEq(t, expectedJSON, w.Body.String())
}

func TestStrictServerGetActivity(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	tests := []struct {
		name         string
		activityId   string
		expectedJSON string
	}{
		{
			name:         "FeedingActivity",
			activityId:   "feeding-1",
			expectedJSON: `{"id":"feeding-1","activityType":"feeding","duration":30,"foodType":"kibble"}`,
		},
		{
			name:         "WalkingActivity",
			activityId:   "walking-1",
			expectedJSON: `{"id":"walking-1","activityType":"walking","duration":60,"distance":2.5}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/pets/pet-1/activities/"+tt.activityId, nil)
			w := httptest.NewRecorder()

			handler.GetActivity(w, req, "pet-1", tt.activityId)

			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

func TestStrictServerLogActivity(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	tests := []struct {
		name         string
		activityJSON string
		expectedJSON string
	}{
		{
			name:         "FeedingActivity",
			activityJSON: `{"activityType": "feeding", "duration": 30, "foodType": "kibble"}`,
			expectedJSON: `{"id": "new-feeding", "activityType": "feeding", "duration": 30, "foodType": "kibble"}`,
		},
		{
			name:         "WalkingActivity",
			activityJSON: `{"activityType": "walking", "duration": 60, "distance": 2.5}`,
			expectedJSON: `{"id": "new-walking", "activityType": "walking", "duration": 60, "distance": 2.5}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/pets/pet-1/activities", bytes.NewBufferString(tt.activityJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.LogActivity(w, req, "pet-1")

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

func TestStrictServerGetHealthRecords(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	req := httptest.NewRequest(http.MethodGet, "/pets/pet-1/health-records", nil)
	w := httptest.NewRecorder()

	handler.GetHealthRecords(w, req, "pet-1")

	expectedJSON := `[
		{"id":"1","recordType":"vaccination","date":"2024-01-15","vaccine":"Rabies"},
		{"id":"2","recordType":"checkup","date":"2024-02-20","weight":4.5}
	]`
	assert.JSONEq(t, expectedJSON, w.Body.String())
}

func TestStrictServerGetHealthRecord(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	tests := []struct {
		name         string
		recordId     string
		expectedJSON string
	}{
		{
			name:         "VaccinationRecord",
			recordId:     "vaccination-1",
			expectedJSON: `{"id":"vaccination-1","recordType":"vaccination","date":"2024-01-15","vaccine":"Rabies"}`,
		},
		{
			name:         "CheckupRecord",
			recordId:     "checkup-1",
			expectedJSON: `{"id":"checkup-1","recordType":"checkup","date":"2024-02-20","weight":4.5}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/pets/pet-1/health-records/"+tt.recordId, nil)
			w := httptest.NewRecorder()

			handler.GetHealthRecord(w, req, "pet-1", tt.recordId)

			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

func TestStrictServerAddHealthRecord(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	tests := []struct {
		name         string
		recordJSON   string
		expectedJSON string
	}{
		{
			name:         "VaccinationRecord",
			recordJSON:   `{"recordType": "vaccination", "date": "2024-01-15", "vaccine": "Rabies"}`,
			expectedJSON: `{"id": "new-vaccination", "recordType": "vaccination", "date": "2024-01-15", "vaccine": "Rabies"}`,
		},
		{
			name:         "CheckupRecord",
			recordJSON:   `{"recordType": "checkup", "date": "2024-02-20", "weight": 4.5}`,
			expectedJSON: `{"id": "new-checkup", "recordType": "checkup", "date": "2024-02-20", "weight": 4.5}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/pets/pet-1/health-records", bytes.NewBufferString(tt.recordJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.AddHealthRecord(w, req, "pet-1")

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

func TestStrictServerGetAllAnimals(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	req := httptest.NewRequest(http.MethodGet, "/animals", nil)
	w := httptest.NewRecorder()

	handler.GetAllAnimals(w, req)

	expectedJSON := `[
		{"id":"1","animalType":"domestic","domesticType":"housecat","name":"Fluffy","owner":"John"},
		{"id":"2","animalType":"wild","wildType":"lion","name":"Simba","habitat":"Savanna"}
	]`
	assert.JSONEq(t, expectedJSON, w.Body.String())
}

func TestStrictServerGetAnimal(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	tests := []struct {
		name         string
		animalId     string
		expectedJSON string
	}{
		{
			name:         "DomesticAnimal",
			animalId:     "domestic-1",
			expectedJSON: `{"id":"domestic-1","animalType":"domestic","domesticType":"housecat","name":"Fluffy","owner":"John"}`,
		},
		{
			name:         "WildAnimal",
			animalId:     "wild-1",
			expectedJSON: `{"id":"wild-1","animalType":"wild","wildType":"lion","name":"Simba","habitat":"Savanna"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/animals/"+tt.animalId, nil)
			w := httptest.NewRecorder()

			handler.GetAnimal(w, req, tt.animalId)

			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

func TestStrictServerRegisterAnimal(t *testing.T) {
	server := &testServer{}
	handler := NewStrictHandler(server, nil)

	tests := []struct {
		name         string
		animalJSON   string
		expectedJSON string
	}{
		{
			name:         "DomesticAnimal",
			animalJSON:   `{"animalType": "domestic", "domesticType": "housecat", "name": "Whiskers", "owner": "John"}`,
			expectedJSON: `{"id": "new-domestic", "animalType": "domestic", "domesticType": "housecat", "name": "Whiskers", "owner": "John"}`,
		},
		{
			name:         "WildAnimal",
			animalJSON:   `{"animalType": "wild", "wildType": "lion", "name": "Simba", "habitat": "Savanna"}`,
			expectedJSON: `{"id": "new-wild", "animalType": "wild", "wildType": "lion", "name": "Simba", "habitat": "Savanna"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/animals", bytes.NewBufferString(tt.animalJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RegisterAnimal(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

type testServer struct {
	StrictServerInterface
}

func (s *testServer) GetPets(_ context.Context, _ GetPetsRequestObject) (GetPetsResponseObject, error) {
	cat1 := Cat{Id: stringPtr("1"), PetType: "cat", Name: "Whiskers", Meow: boolPtr(true)}
	dog := Dog{Id: stringPtr("2"), PetType: "dog", Name: "Buddy", Bark: boolPtr(true)}
	cat2 := Cat{Id: stringPtr("3"), PetType: "cat", Name: "Mittens", Meow: boolPtr(false)}

	return GetPets200PetInterfaceResponse{cat1, dog, cat2}, nil
}

func (s *testServer) GetPet(_ context.Context, request GetPetRequestObject) (GetPetResponseObject, error) {
	switch request.PetId {
	case "cat-1":
		cat := Cat{Id: stringPtr("cat-1"), PetType: "cat", Name: "Whiskers", Meow: boolPtr(true)}
		return cat, nil
	case "dog-1":
		dog := Dog{Id: stringPtr("dog-1"), PetType: "dog", Name: "Buddy", Bark: boolPtr(true)}
		return dog, nil
	default:
		cat := Cat{Id: stringPtr("unknown"), PetType: "cat", Name: "Unknown", Meow: boolPtr(false)}
		return cat, nil
	}
}

func (s *testServer) CreatePet(_ context.Context, request CreatePetRequestObject) (CreatePetResponseObject, error) {
	pet := request.Body
	if pet.IsCat() {
		cat, err := pet.AsCat()
		if err != nil {
			return nil, err
		}
		cat.Id = stringPtr("new-cat")
		return *cat, nil
	} else if pet.IsDog() {
		dog, err := pet.AsDog()
		if err != nil {
			return nil, err
		}
		dog.Id = stringPtr("new-dog")
		return *dog, nil
	}
	return nil, fmt.Errorf("unknown pet type: %s", pet.PetType)
}

func (s *testServer) GetActivities(_ context.Context, _ GetActivitiesRequestObject) (GetActivitiesResponseObject, error) {
	feeding := FeedingActivity{Id: stringPtr("1"), ActivityType: "feeding", Duration: 30, FoodType: "kibble"}
	walking := WalkingActivity{Id: stringPtr("2"), ActivityType: "walking", Duration: 60, Distance: 2.5}

	return GetActivities200PetActivityInterfaceResponse{feeding, walking}, nil
}

func (s *testServer) GetActivity(_ context.Context, request GetActivityRequestObject) (GetActivityResponseObject, error) {
	switch request.ActivityId {
	case "feeding-1":
		feeding := FeedingActivity{Id: stringPtr("feeding-1"), ActivityType: "feeding", Duration: 30, FoodType: "kibble"}
		return feeding, nil
	case "walking-1":
		walking := WalkingActivity{Id: stringPtr("walking-1"), ActivityType: "walking", Duration: 60, Distance: 2.5}
		return walking, nil
	default:
		feeding := FeedingActivity{Id: stringPtr("unknown"), ActivityType: "feeding", Duration: 0, FoodType: "unknown"}
		return feeding, nil
	}
}

func (s *testServer) LogActivity(_ context.Context, request LogActivityRequestObject) (LogActivityResponseObject, error) {
	activity := request.Body

	if activity.IsFeedingActivity() {
		feeding, err := activity.AsFeedingActivity()
		if err != nil {
			return nil, err
		}
		feeding.Id = stringPtr("new-feeding")
		return *feeding, nil
	} else if activity.IsWalkingActivity() {
		walking, err := activity.AsWalkingActivity()
		if err != nil {
			return nil, err
		}
		walking.Id = stringPtr("new-walking")
		return *walking, nil
	}

	return nil, fmt.Errorf("unknown activity type: %s", activity.ActivityType)
}

func (s *testServer) GetHealthRecords(_ context.Context, _ GetHealthRecordsRequestObject) (GetHealthRecordsResponseObject, error) {
	vaccination := VaccinationRecord{
		Id:         stringPtr("1"),
		RecordType: "vaccination",
		Date:       openapi_types.Date{Time: mustParseDate("2024-01-15")},
		Vaccine:    "Rabies",
	}
	checkup := CheckupRecord{
		Id:         stringPtr("2"),
		RecordType: "checkup",
		Date:       openapi_types.Date{Time: mustParseDate("2024-02-20")},
		Weight:     4.5,
	}

	return GetHealthRecords200HealthRecordInterfaceResponse{vaccination, checkup}, nil
}

func (s *testServer) GetHealthRecord(_ context.Context, request GetHealthRecordRequestObject) (GetHealthRecordResponseObject, error) {
	switch request.RecordId {
	case "vaccination-1":
		vaccination := VaccinationRecord{
			Id:         stringPtr("vaccination-1"),
			RecordType: "vaccination",
			Date:       openapi_types.Date{Time: mustParseDate("2024-01-15")},
			Vaccine:    "Rabies",
		}
		return vaccination, nil
	case "checkup-1":
		checkup := CheckupRecord{
			Id:         stringPtr("checkup-1"),
			RecordType: "checkup",
			Date:       openapi_types.Date{Time: mustParseDate("2024-02-20")},
			Weight:     4.5,
		}
		return checkup, nil
	default:
		vaccination := VaccinationRecord{
			Id:         stringPtr("unknown"),
			RecordType: "vaccination",
			Date:       openapi_types.Date{Time: mustParseDate("1970-01-01")},
			Vaccine:    "unknown",
		}
		return vaccination, nil
	}
}

func (s *testServer) AddHealthRecord(_ context.Context, request AddHealthRecordRequestObject) (AddHealthRecordResponseObject, error) {
	record := request.Body

	if record.IsVaccinationRecord() {
		vaccination, err := record.AsVaccinationRecord()
		if err != nil {
			return nil, err
		}
		vaccination.Id = stringPtr("new-vaccination")
		return *vaccination, nil
	} else if record.IsCheckupRecord() {
		checkup, err := record.AsCheckupRecord()
		if err != nil {
			return nil, err
		}
		checkup.Id = stringPtr("new-checkup")
		return *checkup, nil
	}

	return nil, fmt.Errorf("unknown health record type: %s", record.RecordType)
}

func (s *testServer) GetAllAnimals(_ context.Context, _ GetAllAnimalsRequestObject) (GetAllAnimalsResponseObject, error) {
	domestic := DomesticAnimal{Id: stringPtr("1"), AnimalType: "domestic", DomesticType: "housecat", Name: "Fluffy", Owner: stringPtr("John")}
	wild := WildAnimal{Id: stringPtr("2"), AnimalType: "wild", WildType: "lion", Name: "Simba", Habitat: stringPtr("Savanna")}

	return GetAllAnimals200AnimalInterfaceResponse{domestic, wild}, nil
}

func (s *testServer) GetAnimal(_ context.Context, request GetAnimalRequestObject) (GetAnimalResponseObject, error) {
	switch request.AnimalId {
	case "domestic-1":
		domestic := DomesticAnimal{Id: stringPtr("domestic-1"), AnimalType: "domestic", DomesticType: "housecat", Name: "Fluffy", Owner: stringPtr("John")}
		return domestic, nil
	case "wild-1":
		wild := WildAnimal{Id: stringPtr("wild-1"), AnimalType: "wild", WildType: "lion", Name: "Simba", Habitat: stringPtr("Savanna")}
		return wild, nil
	default:
		domestic := DomesticAnimal{Id: stringPtr("unknown"), AnimalType: "domestic", DomesticType: "unknown", Name: "Unknown", Owner: stringPtr("Unknown")}
		return domestic, nil
	}
}

func (s *testServer) RegisterAnimal(_ context.Context, request RegisterAnimalRequestObject) (RegisterAnimalResponseObject, error) {
	animal := request.Body

	if animal.IsDomesticAnimal() {
		domestic, err := animal.AsDomesticAnimal()
		if err != nil {
			return nil, err
		}
		domestic.Id = stringPtr("new-domestic")
		return *domestic, nil
	} else if animal.IsWildAnimal() {
		wild, err := animal.AsWildAnimal()
		if err != nil {
			return nil, err
		}
		wild.Id = stringPtr("new-wild")
		return *wild, nil
	}

	return nil, fmt.Errorf("unknown animal type: %s", animal.AnimalType)
}
