package server

import (
	"net/http"
	"time"

	"github.com/Tanibox/tania-server/config"
	"github.com/Tanibox/tania-server/src/assets/domain"
	"github.com/Tanibox/tania-server/src/assets/query"
	"github.com/Tanibox/tania-server/src/assets/repository"
	"github.com/Tanibox/tania-server/src/assets/storage"
	"github.com/Tanibox/tania-server/src/helper/imagehelper"
	"github.com/Tanibox/tania-server/src/helper/stringhelper"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

// FarmServer ties the routes and handlers with injected dependencies
type FarmServer struct {
	FarmRepo               repository.FarmRepository
	ReservoirRepo          repository.ReservoirRepository
	AreaRepo               repository.AreaRepository
	AreaQuery              query.AreaQuery
	InventoryMaterialRepo  repository.InventoryMaterialRepository
	InventoryMaterialQuery query.InventoryMaterialQuery
	File                   File
}

// NewFarmServer initializes FarmServer's dependencies and create new FarmServer struct
func NewFarmServer(
	farmStorage *storage.FarmStorage,
	areaStorage *storage.AreaStorage,
	reservoirStorage *storage.ReservoirStorage,
	inventoryMaterialStorage *storage.InventoryMaterialStorage,
) (*FarmServer, error) {
	farmRepo := repository.NewFarmRepositoryInMemory(farmStorage)

	areaRepo := repository.NewAreaRepositoryInMemory(areaStorage)
	areaQuery := query.NewAreaQueryInMemory(areaStorage)

	reservoirRepo := repository.NewReservoirRepositoryInMemory(reservoirStorage)

	inventoryMaterialRepo := repository.NewInventoryMaterialRepositoryInMemory(inventoryMaterialStorage)
	inventoryMaterialQuery := query.NewInventoryMaterialQueryInMemory(inventoryMaterialStorage)

	farmServer := FarmServer{
		FarmRepo:               farmRepo,
		ReservoirRepo:          reservoirRepo,
		AreaRepo:               areaRepo,
		AreaQuery:              areaQuery,
		InventoryMaterialRepo:  inventoryMaterialRepo,
		InventoryMaterialQuery: inventoryMaterialQuery,
		File: LocalFile{},
	}

	return &farmServer, nil
}

// Mount defines the FarmServer's endpoints with its handlers
func (s *FarmServer) Mount(g *echo.Group) {
	g.GET("/types", s.GetTypes)
	g.GET("/inventories/plant_types", s.GetInventoryPlantTypes)
	g.GET("/inventories", s.GetAvailableInventories)
	g.POST("/inventories", s.SaveInventory)

	g.POST("", s.SaveFarm)
	g.GET("", s.FindAllFarm)
	g.GET("/:id", s.FindFarmByID)
	g.POST("/:id/reservoirs", s.SaveReservoir)
	g.POST("/reservoirs/:id/notes", s.SaveReservoirNotes)
	g.DELETE("/reservoirs/:reservoir_id/notes/:note_id", s.RemoveReservoirNotes)
	g.GET("/:id/reservoirs", s.GetFarmReservoirs)
	g.GET("/:farm_id/reservoirs/:reservoir_id", s.GetReservoirsByID)
	g.POST("/:id/areas", s.SaveArea)
	g.POST("/areas/:id/notes", s.SaveAreaNotes)
	g.DELETE("/areas/:area_id/notes/:note_id", s.RemoveAreaNotes)
	g.GET("/:id/areas", s.GetFarmAreas)
	g.GET("/:farm_id/areas/:area_id", s.GetAreasByID)
	g.GET("/:farm_id/areas/:area_id/photos", s.GetAreaPhotos)
}

// GetTypes is a FarmServer's handle to get farm types
func (s *FarmServer) GetTypes(c echo.Context) error {
	types := domain.FindAllFarmTypes()

	return c.JSON(http.StatusOK, types)
}

func (s FarmServer) FindAllFarm(c echo.Context) error {
	data := make(map[string][]SimpleFarm)

	result := <-s.FarmRepo.FindAll()
	if result.Error != nil {
		return result.Error
	}

	farms, ok := result.Result.([]domain.Farm)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Internal server error")
	}

	data["data"] = MapToSimpleFarm(farms)

	return c.JSON(http.StatusOK, data)
}

// SaveFarm is a FarmServer's handler to save new Farm
func (s *FarmServer) SaveFarm(c echo.Context) error {
	data := make(map[string]domain.Farm)

	farm, err := domain.CreateFarm(c.FormValue("name"), c.FormValue("farm_type"))
	if err != nil {
		return Error(c, err)
	}

	err = farm.ChangeGeoLocation(c.FormValue("latitude"), c.FormValue("longitude"))
	if err != nil {
		return Error(c, err)
	}

	err = farm.ChangeRegion(c.FormValue("country_code"), c.FormValue("city_code"))
	if err != nil {
		return Error(c, err)
	}

	err = <-s.FarmRepo.Save(&farm)
	if err != nil {
		return Error(c, err)
	}

	data["data"] = farm

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) FindFarmByID(c echo.Context) error {
	data := make(map[string]domain.Farm)

	result := <-s.FarmRepo.FindByID(c.Param("id"))
	if result.Error != nil {
		return result.Error
	}

	farm, ok := result.Result.(domain.Farm)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Internal server error")
	}

	data["data"] = farm

	return c.JSON(http.StatusOK, data)
}

// SaveReservoir is a FarmServer's handler to save new Reservoir and place it to a Farm
func (s *FarmServer) SaveReservoir(c echo.Context) error {
	data := make(map[string]DetailReservoir)
	validation := RequestValidation{}

	// Validate requests //
	name, err := validation.ValidateReservoirName(c.FormValue("name"))
	if err != nil {
		return Error(c, err)
	}

	waterSourceType, err := validation.ValidateType(c.FormValue("type"))
	if err != nil {
		return Error(c, err)
	}

	capacity, err := validation.ValidateCapacity(waterSourceType, c.FormValue("capacity"))
	if err != nil {
		return Error(c, err)
	}

	farm, err := validation.ValidateFarm(*s, c.Param("id"))
	if err != nil {
		return Error(c, err)
	}

	// Process //
	r, err := domain.CreateReservoir(farm, name)
	if err != nil {
		return Error(c, err)
	}

	if waterSourceType == "bucket" {
		b, err := domain.CreateBucket(capacity, 0)
		if err != nil {
			return Error(c, err)
		}

		r.AttachBucket(b)
	} else if waterSourceType == "tap" {
		t, err := domain.CreateTap()
		if err != nil {
			return Error(c, err)
		}

		r.AttachTap(t)
	}

	err = farm.AddReservoir(&r)
	if err != nil {
		return Error(c, err)
	}

	// Persists //
	err = <-s.ReservoirRepo.Save(&r)
	if err != nil {
		return Error(c, err)
	}

	err = <-s.FarmRepo.Save(&farm)
	if err != nil {
		return Error(c, err)
	}

	detailReservoir, err := MapToDetailReservoir(s, r)
	if err != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	data["data"] = detailReservoir

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) SaveReservoirNotes(c echo.Context) error {
	data := make(map[string]DetailReservoir)

	reservoirID := c.Param("id")
	content := c.FormValue("content")

	// Validate //
	result := <-s.ReservoirRepo.FindByID(reservoirID)
	if result.Error != nil {
		return Error(c, result.Error)
	}

	reservoir, ok := result.Result.(domain.Reservoir)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	result = <-s.FarmRepo.FindByID(reservoir.Farm.UID.String())
	if result.Error != nil {
		return Error(c, result.Error)
	}

	farm, ok := result.Result.(domain.Farm)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	if content == "" {
		return Error(c, NewRequestValidationError(REQUIRED, "content"))
	}

	// Process //
	reservoir.AddNewNote(content)
	farm.ChangeReservoirInformation(reservoir)

	// Persists //
	resultSave := <-s.ReservoirRepo.Save(&reservoir)
	if resultSave != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	resultSave = <-s.FarmRepo.Save(&farm)
	if resultSave != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	detailReservoir, err := MapToDetailReservoir(s, reservoir)
	if err != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	data["data"] = detailReservoir

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) RemoveReservoirNotes(c echo.Context) error {
	data := make(map[string]DetailReservoir)

	reservoirID := c.Param("reservoir_id")
	noteID := c.Param("note_id")

	// Validate //
	result := <-s.ReservoirRepo.FindByID(reservoirID)
	if result.Error != nil {
		return Error(c, result.Error)
	}

	reservoir, ok := result.Result.(domain.Reservoir)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	result = <-s.FarmRepo.FindByID(reservoir.Farm.UID.String())
	if result.Error != nil {
		return Error(c, result.Error)
	}

	farm, ok := result.Result.(domain.Farm)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	// Process //
	err := reservoir.RemoveNote(noteID)
	if err != nil {
		return Error(c, err)
	}

	err = farm.ChangeReservoirInformation(reservoir)
	if err != nil {
		return Error(c, err)
	}

	// Persists //
	resultSave := <-s.ReservoirRepo.Save(&reservoir)
	if resultSave != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	resultSave = <-s.FarmRepo.Save(&farm)
	if resultSave != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	detailReservoir, err := MapToDetailReservoir(s, reservoir)
	if err != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	data["data"] = detailReservoir

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) GetFarmReservoirs(c echo.Context) error {
	data := make(map[string][]DetailReservoir)

	result := <-s.FarmRepo.FindByID(c.Param("id"))
	if result.Error != nil {
		return result.Error
	}

	farm, ok := result.Result.(domain.Farm)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Internal server error")
	}

	reservoirs, err := MapToReservoir(s, farm.Reservoirs)
	if err != nil {
		return Error(c, err)
	}

	data["data"] = reservoirs
	if len(farm.Reservoirs) == 0 {
		data["data"] = []DetailReservoir{}
	}

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) GetReservoirsByID(c echo.Context) error {
	data := make(map[string]DetailReservoir)

	// Validate //
	result := <-s.FarmRepo.FindByID(c.Param("farm_id"))
	if result.Error != nil {
		return Error(c, result.Error)
	}

	_, ok := result.Result.(domain.Farm)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusBadRequest, "Internal server error"))
	}

	result = <-s.ReservoirRepo.FindByID(c.Param("reservoir_id"))
	if result.Error != nil {
		return Error(c, result.Error)
	}

	reservoir, ok := result.Result.(domain.Reservoir)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusBadRequest, "Internal server error"))
	}

	detailReservoir, err := MapToDetailReservoir(s, reservoir)
	if err != nil {
		return Error(c, err)
	}

	data["data"] = detailReservoir

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) SaveArea(c echo.Context) error {
	data := make(map[string]DetailArea)
	validation := RequestValidation{}

	// Validation //
	farm, err := validation.ValidateFarm(*s, c.Param("id"))
	if err != nil {
		return Error(c, err)
	}

	reservoir, err := validation.ValidateReservoir(*s, c.FormValue("reservoir_id"))
	if err != nil {
		return Error(c, err)
	}

	size, err := validation.ValidateAreaSize(c.FormValue("size"), c.FormValue("size_unit"))
	if err != nil {
		return Error(c, err)
	}

	location, err := validation.ValidateAreaLocation(c.FormValue("location"))
	if err != nil {
		return Error(c, err)
	}

	// Process //
	area, err := domain.CreateArea(farm, c.FormValue("name"), c.FormValue("type"))
	if err != nil {
		return Error(c, err)
	}

	err = area.ChangeSize(size)
	if err != nil {
		return Error(c, err)
	}

	err = area.ChangeLocation(location)
	if err != nil {
		return Error(c, err)
	}

	photo, err := c.FormFile("photo")
	if err == nil {
		destPath := stringhelper.Join(*config.Config.UploadPathArea, "/", photo.Filename)
		err = s.File.Upload(photo, destPath)

		if err != nil {
			return Error(c, err)
		}

		width, height, err := imagehelper.GetImageDimension(destPath)
		if err != nil {
			return Error(c, err)
		}

		areaPhoto := domain.AreaPhoto{
			Filename: photo.Filename,
			MimeType: photo.Header["Content-Type"][0],
			Size:     int(photo.Size),
			Width:    width,
			Height:   height,
		}

		area.Photo = areaPhoto
	}

	area.Farm = farm
	area.Reservoir = reservoir

	err = farm.AddArea(&area)
	if err != nil {
		return Error(c, err)
	}

	// Persists //
	err = <-s.ReservoirRepo.Save(&reservoir)
	if err != nil {
		return Error(c, err)
	}

	err = <-s.AreaRepo.Save(&area)
	if err != nil {
		return Error(c, err)
	}

	err = <-s.FarmRepo.Save(&farm)
	if err != nil {
		return Error(c, err)
	}

	data["data"] = MapToDetailArea(area)

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) SaveAreaNotes(c echo.Context) error {
	data := make(map[string]DetailArea)

	areaID := c.Param("id")
	content := c.FormValue("content")

	// Validate //
	result := <-s.AreaRepo.FindByID(areaID)
	if result.Error != nil {
		return Error(c, result.Error)
	}

	area, ok := result.Result.(domain.Area)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	result = <-s.FarmRepo.FindByID(area.Farm.UID.String())
	if result.Error != nil {
		return Error(c, result.Error)
	}

	farm, ok := result.Result.(domain.Farm)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	if content == "" {
		return Error(c, NewRequestValidationError(REQUIRED, "content"))
	}

	// Process //
	area.AddNewNote(content)
	farm.ChangeAreaInformation(area)

	// Persists //
	resultSave := <-s.AreaRepo.Save(&area)
	if resultSave != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	resultSave = <-s.FarmRepo.Save(&farm)
	if resultSave != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	data["data"] = MapToDetailArea(area)

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) RemoveAreaNotes(c echo.Context) error {
	data := make(map[string]DetailArea)

	areaID := c.Param("area_id")
	noteID := c.Param("note_id")

	// Validate //
	result := <-s.AreaRepo.FindByID(areaID)
	if result.Error != nil {
		return Error(c, result.Error)
	}

	area, ok := result.Result.(domain.Area)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	result = <-s.FarmRepo.FindByID(area.Farm.UID.String())
	if result.Error != nil {
		return Error(c, result.Error)
	}

	farm, ok := result.Result.(domain.Farm)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	// Process //
	err := area.RemoveNote(noteID)
	if err != nil {
		return Error(c, err)
	}

	err = farm.ChangeAreaInformation(area)
	if err != nil {
		return Error(c, err)
	}

	// Persists //
	resultSave := <-s.AreaRepo.Save(&area)
	if resultSave != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	resultSave = <-s.FarmRepo.Save(&farm)
	if resultSave != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	data["data"] = MapToDetailArea(area)

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) GetFarmAreas(c echo.Context) error {
	data := make(map[string][]domain.Area)

	result := <-s.FarmRepo.FindByID(c.Param("id"))
	if result.Error != nil {
		return Error(c, result.Error)
	}

	farm, ok := result.Result.(domain.Farm)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusBadRequest, "Internal server error"))
	}

	data["data"] = MapToArea(farm.Areas)
	if len(farm.Areas) == 0 {
		data["data"] = []domain.Area{}
	}

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) GetAreasByID(c echo.Context) error {
	data := make(map[string]DetailArea)

	// Validate //
	result := <-s.FarmRepo.FindByID(c.Param("farm_id"))
	if result.Error != nil {
		return Error(c, result.Error)
	}

	_, ok := result.Result.(domain.Farm)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusBadRequest, "Internal server error"))
	}

	result = <-s.AreaRepo.FindByID(c.Param("area_id"))
	if result.Error != nil {
		return Error(c, result.Error)
	}

	area, ok := result.Result.(domain.Area)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusBadRequest, "Internal server error"))
	}

	data["data"] = MapToDetailArea(area)

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) GetAreaPhotos(c echo.Context) error {
	// Validate //
	result := <-s.FarmRepo.FindByID(c.Param("farm_id"))
	if result.Error != nil {
		return Error(c, result.Error)
	}

	_, ok := result.Result.(domain.Farm)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusBadRequest, "Internal server error"))
	}

	result = <-s.AreaRepo.FindByID(c.Param("area_id"))
	if result.Error != nil {
		return Error(c, result.Error)
	}

	area, ok := result.Result.(domain.Area)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusBadRequest, "Internal server error"))
	}

	if area.Photo.Filename == "" {
		return Error(c, NewRequestValidationError(NOT_FOUND, "photo"))
	}

	// Process //
	srcPath := stringhelper.Join(*config.Config.UploadPathArea, "/", area.Photo.Filename)

	return c.File(srcPath)
}

func (s *FarmServer) GetInventoryPlantTypes(c echo.Context) error {
	plantTypes := MapToPlantType(domain.GetPlantTypes())

	return c.JSON(http.StatusOK, plantTypes)
}

func (s *FarmServer) SaveInventory(c echo.Context) error {
	data := make(map[string]InventoryMaterial)

	pType := c.FormValue("plant_type")
	variety := c.FormValue("variety")

	// Validate //
	var plantType domain.PlantType
	switch pType {
	case "vegetable":
		plantType = domain.Vegetable{}
	case "fruit":
		plantType = domain.Fruit{}
	case "herb":
		plantType = domain.Herb{}
	case "flower":
		plantType = domain.Flower{}
	case "tree":
		plantType = domain.Tree{}
	default:
		return Error(c, NewRequestValidationError(NOT_FOUND, "plant_type"))
	}

	// Process //
	inventoryMaterial, err := domain.CreateInventoryMaterial(plantType, variety)
	if err != nil {
		return Error(c, err)
	}

	// Persist //
	err = <-s.InventoryMaterialRepo.Save(&inventoryMaterial)
	if err != nil {
		return Error(c, err)
	}

	data["data"] = MapToInventoryMaterial(inventoryMaterial)

	return c.JSON(http.StatusOK, data)
}

func (s *FarmServer) GetAvailableInventories(c echo.Context) error {
	data := make(map[string][]AvailableInventory)

	// Process //
	result := <-s.InventoryMaterialRepo.FindAll()

	inventories, ok := result.Result.([]domain.InventoryMaterial)
	if !ok {
		return Error(c, echo.NewHTTPError(http.StatusBadRequest, "Internal server error"))
	}

	data["data"] = MapToAvailableInventories(inventories)

	return c.JSON(http.StatusOK, data)
}

func initDataDemo(
	server *FarmServer,
	farmStorage *storage.FarmStorage,
	areaStorage *storage.AreaStorage,
	reservoirStorage *storage.ReservoirStorage,
	inventoryMaterialStorage *storage.InventoryMaterialStorage,
) {
	farmUID, _ := uuid.NewV4()
	farm1 := domain.Farm{
		UID:         farmUID,
		Name:        "MyFarm",
		Type:        "organic",
		Latitude:    "10.00",
		Longitude:   "11.00",
		CountryCode: "ID",
		CityCode:    "JK",
		IsActive:    true,
	}

	farmStorage.FarmMap[farmUID] = farm1

	uid, _ := uuid.NewV4()

	noteUID, _ := uuid.NewV4()
	reservoirNotes := make(map[uuid.UUID]domain.ReservoirNote, 0)
	reservoirNotes[noteUID] = domain.ReservoirNote{
		UID:         noteUID,
		Content:     "Don't forget to close the bucket after using",
		CreatedDate: time.Now(),
	}

	reservoir1 := domain.Reservoir{
		UID:         uid,
		Name:        "MyBucketReservoir",
		PH:          8,
		EC:          12.5,
		Temperature: 29,
		WaterSource: domain.Bucket{Capacity: 100, Volume: 10},
		Farm:        farm1,
		Notes:       reservoirNotes,
		CreatedDate: time.Now(),
	}

	farm1.AddReservoir(&reservoir1)
	farmStorage.FarmMap[farmUID] = farm1
	reservoirStorage.ReservoirMap[uid] = reservoir1

	uid, _ = uuid.NewV4()
	reservoir2 := domain.Reservoir{
		UID:         uid,
		Name:        "MyTapReservoir",
		PH:          8,
		EC:          12.5,
		Temperature: 29,
		WaterSource: domain.Tap{},
		Farm:        farm1,
		Notes:       make(map[uuid.UUID]domain.ReservoirNote),
		CreatedDate: time.Now(),
	}

	farm1.AddReservoir(&reservoir2)
	farmStorage.FarmMap[farmUID] = farm1
	reservoirStorage.ReservoirMap[uid] = reservoir2

	uid, _ = uuid.NewV4()

	noteUID, _ = uuid.NewV4()
	areaNotes := make(map[uuid.UUID]domain.AreaNote, 0)
	areaNotes[noteUID] = domain.AreaNote{
		UID:         noteUID,
		Content:     "This area should only be used for seeding.",
		CreatedDate: time.Now(),
	}

	area1 := domain.Area{
		UID:       uid,
		Name:      "MySeedingArea",
		Size:      domain.SquareMeter{Value: 10},
		Type:      domain.GetAreaType(domain.AreaTypeSeeding),
		Location:  "indoor",
		Photo:     domain.AreaPhoto{},
		Notes:     areaNotes,
		Reservoir: reservoir2,
		Farm:      farm1,
	}

	farm1.AddArea(&area1)
	farmStorage.FarmMap[farmUID] = farm1
	areaStorage.AreaMap[uid] = area1

	uid, _ = uuid.NewV4()
	area2 := domain.Area{
		UID:       uid,
		Name:      "MyGrowingArea",
		Size:      domain.SquareMeter{Value: 100},
		Type:      domain.GetAreaType(domain.AreaTypeGrowing),
		Location:  "outdoor",
		Photo:     domain.AreaPhoto{},
		Notes:     make(map[uuid.UUID]domain.AreaNote),
		Reservoir: reservoir1,
		Farm:      farm1,
	}

	farm1.AddArea(&area2)
	farmStorage.FarmMap[farmUID] = farm1
	areaStorage.AreaMap[uid] = area2

	uid, _ = uuid.NewV4()
	inventory1 := domain.InventoryMaterial{
		UID:       uid,
		PlantType: domain.Vegetable{},
		Variety:   "Bayam Lu Hsieh",
	}

	inventoryMaterialStorage.InventoryMaterialMap[uid] = inventory1

	uid, _ = uuid.NewV4()
	inventory2 := domain.InventoryMaterial{
		UID:       uid,
		PlantType: domain.Vegetable{},
		Variety:   "Tomat Super One",
	}

	inventoryMaterialStorage.InventoryMaterialMap[uid] = inventory2

	uid, _ = uuid.NewV4()
	inventory3 := domain.InventoryMaterial{
		UID:       uid,
		PlantType: domain.Fruit{},
		Variety:   "Apple Rome Beauty",
	}

	inventoryMaterialStorage.InventoryMaterialMap[uid] = inventory3

	uid, _ = uuid.NewV4()
	inventory4 := domain.InventoryMaterial{
		UID:       uid,
		PlantType: domain.Fruit{},
		Variety:   "Orange Sweet Mandarin",
	}

	inventoryMaterialStorage.InventoryMaterialMap[uid] = inventory4
}
