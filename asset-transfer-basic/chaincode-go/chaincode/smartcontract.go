package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Insert struct field in alphabetic order => to achieve determinism across languages
// Considered good practise
type Car struct {
	AssetType           string `json:"AssetType"`
	Car_ID              string `json:"Car_ID"`              // primary key for each car
	Date_of_manufacture string `json:"Date_of_manufacture"` // all date formats YYYYMMDD
	Misc                string `json:"Misc"`                // any other information needed about the car
}
type CarComponent struct {
	AssetType        string `json:"AssetType"`
	Car_Component_ID string `json:"Car_Component_ID"` // primary key
	Car_ID           string `json:"Car_ID"`           // foreign key to car which makes use of this car component
	Fuelcell_ID      string ` json:"Fuelcell_ID"`     // Cost per distance unit of asset in use
	Date_added       int    `json:"Date_added"`       // eg hydrogen, fuel cell, motor etc any component in car you want
	Date_removed     int    `json:"Date_removed"`     // Name of a Supplier would be expanded in more developed implementation
}
type Supplier struct {
	AssetType     string `json:"AssetType"`
	Supplier_ID   string `json:"Supplier_ID"` // Has the cost incurred from this Journey been paid
	Supplier_name string `json:"Supplier_name"`
	Misc          string `json:"Misc"` //other info not processed
}

type FuelcellData struct {
	AssetType     string  `json:"AssetType"`
	Fuelcell_ID   string  `json:"Fuelcell_ID"` // foreign key for CarComponents
	Supplier_ID   string  `json:"Supplier_ID"`
	Base_rate     float32 `json:"Base_rate"`
	Distance_rate float32 `json:"Distance_rate"`
	Energy_rate   float32 `json:"energy_rate"` //rate of energy consumed per gramme h2 used (modulo efficiency) energy produced?
	Currency      string  `json:"Currency"`
	Date_Received int     `json:"Date_Received"`
	Date_Returned int     `json:"Date_Returned"`
	Misc          string  `json:"Misc"`
}
type JourneyData struct {
	AssetType        string  `json:"AssetType"`
	Journey_ID       string  `json:"Journey_ID"` // primary key for the database
	Car_ID           string  `json:"Car_ID"`     // foreign key for CarComponents // possibly rendered irrelivent using getHistoryForKey method?
	Car_Component_ID string  `json:"Car_Component_ID"`
	Odo_start        int     `json:"Odo_start"`
	Distance         int     `json:"Distance"` // working off rounded to nearest int
	H2_used          int     `json:"H2_used"`
	Efficiency       float32 `json:"Efficiency"`   // range of 0-1 how Efficienct the fuel cell was for the journey
	Journey_date     int     `json:"Journey_date"` // what date did this journey begin
	Billed           bool    `json:"Billed"`
	Misc             string  `json:"Misc"`
}
type Bill struct {
	AssetType   string  `json:"AssetType"`
	Bill_ID     string  `json:"Bill_ID"`     // foreign key for CarComponents
	Supplier_ID string  `json:"Supplier_ID"` // working off rounded to nearest int
	Fuelcell_ID string  `json:"Fuelcell_ID"` // range of 0-1 how Efficienct the vehicle was used in fuel consumption calculation
	Date_from   int     `json:"Date_from"`   // will be passed through to cost calculations as component name for fuel.
	Date_to     int     `json:"Date_to"`     // primary key for the database
	Currency    string  `json:"Currency"`    // what date did this journey begin
	Amount      float32 `json:"Amount"`      // Has the cost incurred from this Journey been paid
}

// InitLedger adds a base set of all assets to the ledger allowing for basic testing
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	Cars := []Car{
		{AssetType: "Car", Car_ID: "Car1", Date_of_manufacture: "20200123", Misc: "the first test car"},
		{AssetType: "Car", Car_ID: "Car2", Date_of_manufacture: "20210123", Misc: "the first car to swap cells"},
		{AssetType: "Car", Car_ID: "Car3", Date_of_manufacture: "20220123", Misc: "the first loaned out car"},
	}

	for _, asset := range Cars {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.Car_ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	CarComponents := []CarComponent{
		{AssetType: "Car_Component", Car_Component_ID: "Component1", Car_ID: "Car1", Fuelcell_ID: "FuelCell1", Date_added: 20200123, Date_removed: 0}, // 0 signifies still in place
		{AssetType: "Car_Component", Car_Component_ID: "Component2", Car_ID: "Car2", Fuelcell_ID: "FuelCell2", Date_added: 20210123, Date_removed: 20210623},
		{AssetType: "Car_Component", Car_Component_ID: "Component3", Car_ID: "Car3", Fuelcell_ID: "FuelCell2", Date_added: 20210623, Date_removed: 0},
		{AssetType: "Car_Component", Car_Component_ID: "Component4", Car_ID: "Car2", Fuelcell_ID: "FuelCell3", Date_added: 20210623, Date_removed: 0}, // possibly redundent if GetHistoryForKey works as expected
	}

	for _, asset := range CarComponents {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.Car_Component_ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	Suppliers := []Supplier{
		{AssetType: "Supplier", Supplier_ID: "Supplier1", Supplier_name: "Hydrogen1", Misc: "Non preferred provider"},
		{AssetType: "Supplier", Supplier_ID: "Supplier2", Supplier_name: "EfficentCells", Misc: "preferred provider"},
	}

	for _, asset := range Suppliers {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState((asset.Supplier_ID), assetJSON) ////////may cause errors
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	Fuelcells := []FuelcellData{
		{AssetType: "Fuelcell", Fuelcell_ID: "FuelCell1", Supplier_ID: "Supplier1", Base_rate: 1, Distance_rate: 0.2, Energy_rate: 1, Currency: "Pounds", Date_Received: 20200122, Date_Returned: 0, Misc: "Test cell not for production"},
		{AssetType: "Fuelcell", Fuelcell_ID: "FuelCell2", Supplier_ID: "Supplier2", Base_rate: 0.5, Distance_rate: 0.1, Energy_rate: 1, Currency: "Pounds", Date_Received: 20210122, Date_Returned: 0},
		{AssetType: "Fuelcell", Fuelcell_ID: "FuelCell3", Supplier_ID: "Supplier2", Base_rate: 0.8, Distance_rate: 0.2, Energy_rate: 1, Currency: "Pounds", Date_Received: 20210622, Date_Returned: 0},
	}

	for _, asset := range Fuelcells {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState((asset.Fuelcell_ID), assetJSON) ////////may cause errors
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	Journeys := []JourneyData{
		{AssetType: "Journey", Journey_ID: "Journey1", Car_ID: "Car1", Car_Component_ID: "Component1", Odo_start: 0, Distance: 500, H2_used: 100, Efficiency: 0.3, Journey_date: 20200123, Billed: false},
		{AssetType: "Journey", Journey_ID: "Journey2", Car_ID: "Car1", Car_Component_ID: "Component1", Odo_start: 500, Distance: 1000, H2_used: 150, Efficiency: 0.5, Journey_date: 20200127, Billed: false},
		{AssetType: "Journey", Journey_ID: "Journey3", Car_ID: "Car2", Car_Component_ID: "Component2", Odo_start: 0, Distance: 500, H2_used: 100, Efficiency: 0.3, Journey_date: 20210127, Billed: false},
		{AssetType: "Journey", Journey_ID: "Journey4", Car_ID: "Car3", Car_Component_ID: "Component3", Odo_start: 0, Distance: 500, H2_used: 100, Efficiency: 0.3, Journey_date: 20210627, Billed: false},
		{AssetType: "Journey", Journey_ID: "Journey5", Car_ID: "Car2", Car_Component_ID: "Component4", Odo_start: 500, Distance: 500, H2_used: 100, Efficiency: 0.3, Journey_date: 20210627, Billed: false},
		{AssetType: "Journey", Journey_ID: "Journey6", Car_ID: "Car1", Car_Component_ID: "Component1", Odo_start: 10000, Distance: 500, H2_used: 100, Efficiency: 0.3, Journey_date: 20210627, Billed: false},
	}

	for _, asset := range Journeys {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState((asset.Journey_ID), assetJSON) ////////may cause errors
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	Bills := []Bill{ //example bill structures
		//{AssetType: "Bill", Bill_ID: "Bill1", Supplier_ID: "Hydrogen1", Fuelcell_ID: "FuelCell1", Date_from: 20200101, Date_to: 20200131},
		//{AssetType: "Bill", Bill_ID: "Bill2", Supplier_ID: "EfficentCells", Fuelcell_ID: "Fuelcells2", Date_from: 20200101, Date_to: 20200101},
	}

	for _, asset := range Bills {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState((asset.Bill_ID), assetJSON) ////////may cause errors
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}
func (s *SmartContract) GenerateBill(ctx contractapi.TransactionContextInterface, bill_ID string, fuelcell_ID string, startDate string, endDate string) error {
	PastBills, err := s.GetAllBills(ctx)
	if err != nil {
		return err
	}
	fmt.Println(("GettingFuelcell"))
	Fuelcell, err := s.GetFuelcell(ctx, fuelcell_ID)
	if err != nil {
		return err
	}
	fmt.Println("GotFuelcell", Fuelcell.Fuelcell_ID)
	intEndDate, err := strconv.Atoi(endDate)
	if err != nil {
		return err
	}
	intStartDate, err := strconv.Atoi(startDate)
	if err != nil {
		return err
	}
	for _, PastBill := range PastBills {
		if PastBill.Supplier_ID == Fuelcell.Supplier_ID {
			if intStartDate < PastBill.Date_from && PastBill.Date_from < intEndDate {
				return fmt.Errorf("this bill would start encapulate another bills start time")
			}
			if intStartDate <= PastBill.Date_to && PastBill.Date_to <= intEndDate {
				return fmt.Errorf("this bill would start encapulate another bills end time")

			}
			if PastBill.Date_from < intStartDate && intStartDate < PastBill.Date_to {
				return fmt.Errorf("this bill would start inside a time frame covered by another bill")
			}
			if PastBill.Date_from <= intEndDate && intEndDate <= PastBill.Date_to {
				return fmt.Errorf("this bill would end inside a time frame covered by another bill")

			}
		}
	}

	var TotalJourneysCost float32
	if Fuelcell.Date_Received > intEndDate { // if fuelcell was recieved after the end date
		return fmt.Errorf("fuelcell %s didn't exist in this time frame", fuelcell_ID)
	}
	if (Fuelcell.Date_Returned != 0) && (Fuelcell.Date_Returned < intStartDate) { // if the fuel cell has been returned and was returned before the start date
		return fmt.Errorf("fuelcell %s had been returned before this time frame", fuelcell_ID)
	}
	// query to get carCompenents relevent, may return multiple if fuelcell moved inside billing time
	fmt.Println(("Getting relevent car components"))
	releventCarComponents, err := s.GetAllCarCompForFuelCellBetweenDates(ctx, Fuelcell.Fuelcell_ID, startDate, endDate)
	fmt.Println(("got relevent car components"))
	if err != nil {
		return err
	}
	// find relevent journey where car_Component is correct and journey is in range of car component installed
	for _, currentCarComponent := range releventCarComponents {
		fmt.Println("GetAllJourneysbetweendatesforCarComponent", currentCarComponent.Car_Component_ID)
		releventJourneys, err := s.GetAllJourneysbetweendatesforCarComponent(ctx, currentCarComponent.Car_Component_ID, startDate, endDate)
		if err != nil {
			return err
		}
		fmt.Println("got", currentCarComponent.Car_Component_ID, "journeys")
		for _, currentJourney := range releventJourneys {
			fmt.Println("checking if current journey has been billed")
			if !currentJourney.Billed {
				fmt.Println("Journey wasnt Billed so doing maths on it")
				JourneyCost := (float32(currentJourney.Distance) * Fuelcell.Distance_rate) + (float32(currentJourney.H2_used) * currentJourney.Efficiency * Fuelcell.Energy_rate)
				TotalJourneysCost += JourneyCost
			}
		}

	}
	fmt.Println("calcuating final cost")
	basecharge := ((float32(intEndDate-intStartDate) + 1) * Fuelcell.Base_rate) // plus one to account for 31st - 1 equalling 30 not 31 days
	TotalCost := basecharge + TotalJourneysCost
	// use the H2, effiency and distance from journey and Baserate, distance rate and energy rate from fuel cell to generate bill cost and create bill
	fmt.Println("creating and submitting the bill")
	error := s.CreateNewBill(ctx, bill_ID, Fuelcell.Supplier_ID, Fuelcell.Fuelcell_ID, startDate, endDate, "£", TotalCost)
	if error != nil {
		return error
	}
	fmt.Println("marking now billed journeys as so billed this does not mean paid")
	for _, currentCarComponent := range releventCarComponents {
		releventJourneys, err := s.GetAllJourneysbetweendatesforCarComponent(ctx, currentCarComponent.Car_Component_ID, startDate, endDate)
		if err != nil {
			return err
		}
		for _, currentJourney := range releventJourneys {
			if !currentJourney.Billed {
				currentJourney.Billed = true
				BilledJourney, err := json.Marshal(currentJourney)
				if err != nil {
					return err
				}
				fmt.Println("submitting the updated journeys as billed")
				err = ctx.GetStub().PutState(currentJourney.Journey_ID, BilledJourney)
				if err != nil {
					return fmt.Errorf("failed to mark Journey %s as billed error: %v", currentJourney.Journey_ID, err)
				}
			}
		}

	}
	return error // successful completion
}

// Get functions used specifically for billing
func (s *SmartContract) GetFuelcell(ctx contractapi.TransactionContextInterface, Fuelcell_ID string) (*FuelcellData, error) {
	queryString := fmt.Sprintf(`{"selector":{"AssetType":"Fuelcell","Fuelcell_ID":"%s"}}`, Fuelcell_ID)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString) // should only return one
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets *FuelcellData
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset FuelcellData
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = &asset
	}

	return assets, nil
}
func (s *SmartContract) GetAllCarCompForFuelCellBetweenDates(ctx contractapi.TransactionContextInterface, FuelcellID string, startDate string, endDate string) ([]*CarComponent, error) {
	queryString := fmt.Sprintf(`{"selector":{"AssetType":"Car_Component","Fuelcell_ID":"%s"}}`, FuelcellID)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*CarComponent
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset CarComponent
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		intEndDate, err := strconv.Atoi(endDate)
		if err != nil {
			return nil, err
		}
		intStartDate, err := strconv.Atoi(startDate)
		if err != nil {
			return nil, err
		}
		fmt.Println("Successfully generated int versons of start string")
		fmt.Println(asset.Car_Component_ID)
		fmt.Println("(asset.Date_removed == 0" + strconv.FormatBool((asset.Date_removed == 0)))
		fmt.Println("or")
		fmt.Println("intStartDate <= asset.Date_removed", strconv.FormatBool(intStartDate <= asset.Date_removed), ")")
		fmt.Println("and")
		fmt.Println("asset.Date_added <= intEndDate", strconv.FormatBool(asset.Date_added <= intEndDate))
		if (asset.Date_removed == 0 || intStartDate <= asset.Date_removed) && (asset.Date_added <= intEndDate) { // (if hasnt been removed or removed after start date) and was installed before the end date
			assets = append(assets, &asset)
		}
	}

	return assets, nil
}
func (s *SmartContract) GetAllJourneysbetweendatesforCarComponent(ctx contractapi.TransactionContextInterface, Car_Component_ID string, startDate string, endDate string) ([]*JourneyData, error) {
	queryString := fmt.Sprintf(`{"selector":{"AssetType":"Journey","Car_Component_ID":"%s"}}`, Car_Component_ID)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*JourneyData
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset JourneyData
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		intStartDate, err := strconv.Atoi(startDate)
		if err != nil {
			return nil, err
		}
		intEndDate, err := strconv.Atoi(endDate)
		if err != nil {
			return nil, err
		}
		if (intStartDate < asset.Journey_date) && (asset.Journey_date < intEndDate) { // as long as journey happened when component was present in those dates
			assets = append(assets, &asset)
		}
	}

	return assets, nil
}

// create new asset functions:
func (s *SmartContract) CreateNewBill(ctx contractapi.TransactionContextInterface, Bill_ID string, Supplier_ID string,
	Fuelcell_ID string, startDate string, endDate string, Currency string, amount float32) error {
	exists, err := s.AssetExists(ctx, Bill_ID) // does the journey already exist
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the bill %s already exists", Bill_ID)
	}
	exists, err = s.AssetExists(ctx, Supplier_ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the Supplier %s Doesn't exist", Supplier_ID)
	}
	exists, err = s.AssetExists(ctx, Fuelcell_ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the FuelCell_ID %s Doesn't exist", Fuelcell_ID)
	}
	intEndDate, err := strconv.Atoi(endDate)
	if err != nil {
		return err
	}
	intStartDate, err := strconv.Atoi(startDate)
	if err != nil {
		return err
	}
	asset := Bill{
		AssetType:   "Bill",
		Bill_ID:     Bill_ID,
		Supplier_ID: Supplier_ID,
		Fuelcell_ID: Fuelcell_ID,
		Date_from:   intStartDate,
		Date_to:     intEndDate,
		Currency:    Currency,
		Amount:      amount,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(Bill_ID, assetJSON)
}

// needs testing
func (s *SmartContract) CreateJourney(ctx contractapi.TransactionContextInterface, Journey_ID string, Car_ID string,
	Car_Component_ID string, Odo_start string, Distance string, H2_used string, Efficiency float32,
	FuelSupplier string, Date string) error {
	exists, err := s.AssetExists(ctx, Journey_ID) // does the journey already exist
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", Journey_ID)
	}
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the Car %s Doesn't exist", Car_ID)
	}
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the FuelSupplier %s Doesn't exist", FuelSupplier)
	}
	intOdo_start, err := strconv.Atoi(Odo_start)
	if err != nil {
		return err
	}
	intDistance, err := strconv.Atoi(Distance)
	if err != nil {
		return err
	}
	intH2_used, err := strconv.Atoi(H2_used)
	if err != nil {
		return err
	}
	intDate, err := strconv.Atoi(Date)
	if err != nil {
		return err
	}
	asset := JourneyData{
		AssetType:        "Journey",
		Journey_ID:       Journey_ID,
		Car_ID:           Car_ID,
		Car_Component_ID: Car_Component_ID,
		Odo_start:        intOdo_start,
		Distance:         intDistance,
		H2_used:          intH2_used,
		Efficiency:       Efficiency,
		Journey_date:     intDate,
		Billed:           false,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(Journey_ID, assetJSON)
}

// Get all of a certain asset functions:
func (s *SmartContract) GetAllJourneys(ctx contractapi.TransactionContextInterface) ([]*JourneyData, error) {
	queryString := `{"selector":{"AssetType":"Journey"}}`
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*JourneyData
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset JourneyData
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
func (s *SmartContract) GetAllCars(ctx contractapi.TransactionContextInterface) ([]*Car, error) {
	queryString := `{"selector":{"AssetType":"Car"}}`
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*Car
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset Car
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
func (s *SmartContract) GetAllCarComponents(ctx contractapi.TransactionContextInterface) ([]*CarComponent, error) {
	queryString := `{"selector":{"AssetType":"Car_Component"}}`
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*CarComponent
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset CarComponent
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
func (s *SmartContract) GetAllSuppliers(ctx contractapi.TransactionContextInterface) ([]*Supplier, error) {
	queryString := `{"selector":{"AssetType":"Supplier"}}`
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*Supplier
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset Supplier
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
func (s *SmartContract) GetAllFuelcells(ctx contractapi.TransactionContextInterface) ([]*FuelcellData, error) {
	queryString := `{"selector":{"AssetType":"Fuelcell"}}`
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*FuelcellData
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset FuelcellData
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
func (s *SmartContract) GetAllBills(ctx contractapi.TransactionContextInterface) ([]*Bill, error) {
	queryString := `{"selector":{"AssetType":"Bill"}}`
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*Bill
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset Bill
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// EXTRA FUNCTIONS PROVIDING FUNCTIONALLITY NOT CURRENTLY UTILISED#########################################################################################
// function for if you wanted all Suppliers FuelCells
func (s *SmartContract) GetAllSuppliersFuelCellsBetweenDates(ctx contractapi.TransactionContextInterface, FuelCell string, startDate string, endDate string) ([]*FuelcellData, error) {
	queryString := fmt.Sprintf(`{"selector":{"AssetType":"Fuelcell","Supplier":"%s"}}`, FuelCell)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*FuelcellData
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset FuelcellData
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		intEndDate, err := strconv.Atoi(endDate)
		if err != nil {
			return nil, err
		}
		intStartDate, err := strconv.Atoi(startDate)
		if err != nil {
			return nil, err
		}
		if asset.Date_Received < intEndDate { // received before the end of the period
			if (asset.Date_Returned == 0) || !(asset.Date_Returned < intStartDate) { // hasn't been returned or wasnt returned before the start of the period
				assets = append(assets, &asset)
			}
		}
	}

	return assets, nil
}
func (s *SmartContract) GetJourneysbyCar(ctx contractapi.TransactionContextInterface, Car_ID string) ([]*JourneyData, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*JourneyData
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset JourneyData
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		if asset.Car_ID == Car_ID {
			assets = append(assets, &asset)
		}
	}
	return assets, nil
}
func (s *SmartContract) GetAllJourneysofCar(ctx contractapi.TransactionContextInterface, term string) ([]*JourneyData, error) {
	queryString := fmt.Sprintf(`{"selector":{"AssetType":"Journey","Car_ID":"%s"}}`, term)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*JourneyData
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset JourneyData
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
func (s *SmartContract) GetAllJourneysbetweendates(ctx contractapi.TransactionContextInterface, startDate string, endDate string) ([]*JourneyData, error) {
	queryString := `{"selector":{"AssetType":"Journey"}}`
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*JourneyData
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset JourneyData
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		intStartDate, err := strconv.Atoi(startDate)
		if err != nil {
			return nil, err
		}
		intEndDate, err := strconv.Atoi(endDate)
		if err != nil {
			return nil, err
		}
		if (intStartDate < asset.Journey_date) && (asset.Journey_date < intEndDate) {
			assets = append(assets, &asset)
		}
	}

	return assets, nil
}

// get contents of a specified journey
func (s *SmartContract) ReadJourney(ctx contractapi.TransactionContextInterface, id string) (*JourneyData, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset JourneyData
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, Asset_ID string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(Asset_ID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}
