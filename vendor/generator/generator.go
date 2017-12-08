package generator

import (
	"os"
	"fmt"
	. "github.com/dave/jennifer/jen"
	"database"
	"log"
)

type Entity struct {
	ID          int `sql:"AUTO_INCREMENT"`
	Name        string `sql:"type:varchar(30)"  gorm:"column:name;not null;unique"`
	DisplayName string `sql:"type:varchar(30)" gorm:"column:display_name"`
	Columns     []Column `gorm:"ForeignKey:entity_id;AssociationForeignKey:id"` // one to many, has many columns
}

type ColumnType struct {
	ID      int    `sql:"AUTO_INCREMENT"`
	Type    string `sql:"type:varchar(30)"`
	Columns []Column `gorm:"ForeignKey:type_id;AssociationForeignKey:id"` //one to many, has many columns
}

type Column struct {
	ID          int `sql:"AUTO_INCREMENT"`
	Name        string `sql:"type:varchar(30)" gorm:"unique_index:idx_name_entity_id"`
	DisplayName string `sql:"type:varchar(30)"`
	Size        int `sql:"type:int(30)"`
	TypeID      int `sql:"type:int(30)"`
	EntityID    int `sql:"type:int(100)" gorm:"unique_index:idx_name_entity_id"`
	ColumnType  ColumnType `gorm:"ForeignKey:TypeID"` //belong to (for reverse access)
}

type RelationType struct {
	ID   int `sql:"AUTO_INCREMENT"`
	Name string `sql:"type:varchar(30)"`
}

type Relation struct {
	ID                int `sql:"AUTO_INCREMENT"`
	ParentEntityID    int `sql:"type:int(100)" gorm:"unique_index:idx_all_relation"`
	ParentEntityColID int `sql:"type:int(100)" gorm:"unique_index:idx_all_relation"`
	ChildEntityID     int `sql:"type:int(100)" gorm:"unique_index:idx_all_relation"`
	ChildEntityColID  int `sql:"type:int(100)" gorm:"unique_index:idx_all_relation"`
	InterEntityID     int `sql:"type:int(100)" gorm:"unique_index:idx_all_relation"`
	RelationTypeID    int `sql:"type:int(10)" gorm:"unique_index:idx_all_relation"`

	ParentEntity Entity `gorm:"ForeignKey:ParentEntityID"`       //belong to
	ChildEntity  Entity `gorm:"ForeignKey:ChildEntityID"`        //belong to
	InterEntity  Entity `gorm:"ForeignKey:InterEntityID"`        //belong to
	ParentColumn Column `gorm:"ForeignKey:ParentEntityColID"`    //belong to
	ChildColumn  Column `gorm:"ForeignKey:ChildEntityColID"`     //belong to
	RelationType RelationType `gorm:"ForeignKey:RelationTypeID"` //belong to
}

func (Entity) TableName() string {
	return "c_entity"
}

func (ColumnType) TableName() string {
	return "c_column_type"
}

func (Column) TableName() string {
	return "c_column"
}

func (RelationType) TableName() string {
	return "c_relation_type"
}

func (Relation) TableName() string {
	return "c_relation"
}

func GenerateCode(appName string) {

	//fetch all entities
	entities := []Entity{}
	database.SQL.Preload("Columns.ColumnType").
		Find(&entities)

	//print all entities
	//for _, entity := range entities {
	//	fmt.Print(entity.Name + " (" + entity.DisplayName + ")\n")
	//	for _, col := range entity.Columns {
	//		fmt.Print("\t", col.Name, " ", col.ColumnType.Type, "(", col.Size, ")\n")
	//	}
	//}

	allModels := make([]string, 0)
	//creating entity structures
	//for _, entity := range entities {
	//	allModels = append(allModels, createEntities(entity, db))
	//}

	//create xShowroom.go
	file, err := os.Create(appName + ".go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	//created file
	appMain := NewFile("main")

	//write all code
	createAppMain(appMain, allModels)

	//flush xShowroom.go
	fmt.Fprintf(file, "%#v", appMain)
	fmt.Println("=========================")
	fmt.Println(appName, "generated!!!")
}

//xShowroom generation methods
func createAppMain(appMain *File, allModels []string) {

	createAppMainInitMethod(appMain)

	createAppMainMainMethod(appMain, allModels)
}

func createAppMainInitMethod(appMain *File) {
	//add init method in appMain.go
	appMain.Func().Id("init").Params().Block(
		Comment(" Use all cpu cores"),
		Qual("runtime", "GOMAXPROCS").Call(Qual("runtime", "NumCPU").Call()),
	)
}

func createAppMainMainMethod(appMain *File, allModels []string) {

	//add main method in appMain.go
	appMain.Func().Id("main").Params().Block(

		Comment("Load the configuration file"),
		Qual("jsonconfig", "Load").Call(
			Lit("config").
				Op("+").
				Id("string").
				Op("(").
				Qual("os", "PathSeparator").
				Op(")").
				Op("+").
				Lit("config.json"),
			Id("config")),

		Empty(),

		Comment("Connect to database"),
		Qual("database", "Connect").Call(
			Id("config").Op(".").Id("Database"),
		),

		Empty(),

		Comment("Load the controller routes"),
		Qual("models", "Load").Call(),

		Empty(),

		Comment("Auto migrate all models"),
		Qual("database", "SQL.AutoMigrate").CallFunc(func(g *Group) {
			for _, value := range allModels {
				g.Id("&" + "models." + value + "{}")
			}
		}),

		Empty(),

		//Comment("Start the listener"),
		//Qual("shared/server", "Run").Call(
		//	Qual("route", "LoadHTTP").Call(),
		//	Qual("route", "LoadHTTPS").Call(),
		//	Id("config").Op(".").Id("Server"),
		//),
	)
}
