package generator

import (
	"os"
	"fmt"
	. "github.com/dave/jennifer/jen"
	"database"
	"log"
	"github.com/jinzhu/gorm"
	"strings"
	"bytes"
	"strconv"
)

var const_ConfigPath = "config"
var const_JsonConfigPath = "jsonconfig"
var const_DatabasePath = "database"
var const_ModelsPath = "models"
var const_ServerPath = "server"
var const_RoutePath = "route"
var const_RouterPath = "router"
var const_OneToOne = "OneToOne"
var const_OneToMany = "OneToMany"
var const_ManyToOne = "ManyToOne"
var const_ManyToMany = "ManyToMany"
var const_reverse = "_reverse"
var const_normal = "_normal"
var const_self = "_self"

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

type EntityRelation struct {
	Type             string
	SubEntityName    string
	SubEntityColName string
}

type EntityRelationMethod struct {
	MethodName       string
	Type             string
	SubEntityName    string
	SubEntityColName string
}

type EntityField struct {
	FieldName string
	FieldType string
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
	for _, entity := range entities {
		allModels = append(allModels, createEntities(entity, database.SQL))
	}

	//create appName.go
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

	//create an instance of configuration
	appMain.Var().Id("conf").Op("= &").Qual("config", "Configuration{}")

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
		Qual(const_JsonConfigPath, "Load").Call(
			Lit(const_ConfigPath).
				Op("+").
				Id("string").
				Op("(").
				Qual("os", "PathSeparator").
				Op(")").
				Op("+").
				Lit("config.json"),
			Id("conf")),

		Empty(),

		Comment("Connect to database"),
		Qual(const_DatabasePath, "Connect").Call(
			Id("conf").Op(".").Id("Database"),
		),

		Empty(),

		Comment("Load the controller routes"),
		Qual(const_ModelsPath, "Load").Call(),

		Empty(),

		Comment("Auto migrate all models"),
		Qual(const_DatabasePath, "SQL.AutoMigrate").CallFunc(func(g *Group) {
			for _, value := range allModels {
				g.Id("&" + "models." + value + "{}")
			}
		}),

		Empty(),

		Comment("Start the listener"),
		Qual(const_ServerPath, "Run").Call(
			Qual(const_RoutePath, "LoadHTTP").Call(),
			Qual(const_RoutePath, "LoadHTTPS").Call(),
			Id("conf").Op(".").Id("Server"),
		),
	)
}

//models generation methods
func createEntities(entity Entity, db *gorm.DB) string {

	// create entity name from table
	entityName := snakeCaseToCamelCase(entity.DisplayName)

	//entity relations stored to generate routes and their methods for each sub entities ((parent to child) and (child to parent))
	entityRelationsForEachEndpoint := []EntityRelation{}

	//entity relations stored to generate one route to access all sub entities depending on query params(parent to child only)
	entityRelationsForAllEndpoint := []EntityRelation{}

	//create entity file in models sub directory
	file, err := os.Create("vendor/models/" + entityName + ".go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	//set package as "models"
	modelFile := NewFile("models")

	//fetch relations of this entity matching parent
	relationsParent := []Relation{}
	db.Preload("InterEntity").
		Preload("ChildEntity").
		Preload("ChildColumn").
		Preload("ParentColumn").
		Where("parent_entity_id=?", entity.ID).
		Find(&relationsParent)

	//fetch relations of this entity matching child
	relationsChild := []Relation{}
	db.Preload("InterEntity").
		Preload("ParentEntity").
		Preload("ChildColumn").
		Preload("ParentColumn").
		Where("child_entity_id=?", entity.ID).
		Find(&relationsChild)

	entityFields := []EntityField{}

	//write structure for entity
	modelFile.Type().Id(entityName).StructFunc(func(g *Group) {

		//write primitive fields
		for _, column := range entity.Columns {
			entityFields = append(entityFields, mapColumnTypes(column, g))
		}

		//write composite fields while looking at parent
		for _, relation := range relationsParent {
			//fmt.Println("parent ", relation)
			name := snakeCaseToCamelCase(relation.ChildEntity.DisplayName)
			childName := string(relation.ChildColumn.Name)
			parentName := string(relation.ParentColumn.Name)

			d := " "
			relType := "_normal"
			if entityName == name {
				d = "*" //if name and entityName are same, its a self join, so add *
				relType = "_self"
			}

			switch relation.RelationTypeID {
			case 1: //one to one
				relationName := name
				finalId := relationName + " " + d + name + " `gorm:\"ForeignKey:" + childName + ";AssociationForeignKey:" + parentName + "\" json:\"" + relation.ChildEntity.DisplayName + ",omitempty\"`"
				entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{"OneToOne" + relType, name, childName})
				entityRelationsForAllEndpoint = append(entityRelationsForAllEndpoint, EntityRelation{"OneToOne" + relType, relationName, childName})
				g.Id(finalId)
			case 2: //one to many
				relationName := name + "s"
				finalId := relationName + " []" + name + " `gorm:\"ForeignKey:" + childName + ";AssociationForeignKey:" + parentName + "\" json:\"" + relation.ChildEntity.DisplayName + "s,omitempty\"`"
				entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{"OneToMany", name, childName})
				entityRelationsForAllEndpoint = append(entityRelationsForAllEndpoint, EntityRelation{"OneToMany", relationName, childName})
				g.Id(finalId)
			case 3: //many to many
				relationName := name + "s"
				finalId := relationName + " []" + name + " `gorm:\"many2many:" + relation.InterEntity.Name + "\" json:\"" + relation.ChildEntity.DisplayName + "s,omitempty\"`"
				g.Id(finalId)
				entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{"ManyToMany", name, childName})
			}
		}

		//write composite fields while looking at child
		for _, relation := range relationsChild {
			name := snakeCaseToCamelCase(relation.ParentEntity.DisplayName)
			childName := string(relation.ChildColumn.Name)

			switch relation.RelationTypeID {
			case 1: //ont to one
				// means current entity's one item belongs to
				if name != entityName { // if check to exclude self join
					entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{const_OneToOne + const_reverse, name, childName})
				}
			case 2: //one to many
				// means current entity's many items belongs to
				finalId := name + " " + name + " `gorm:\"ForeignKey:" + snakeCaseToCamelCase(childName) + "\" json:\"" + name + ",omitempty\"`"
				entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{const_ManyToOne, name, childName})
				g.Id(finalId)
			case 3: //many to many
				// add two record in relation table to create many to many or uncomment this and add relation here
				//fmt.Println("\t\t many to many " + relation.InterEntity.DisplayName + " for " + entityName + " from child")
			}
		}
	})

	//write table name method for our struct
	modelFile.Func().Params(Id(snakeCaseToCamelCase(entity.DisplayName))).Id("TableName").Params().String().Block(
		Return(Lit(entity.Name)),
	)

	getAllMethodName := "GetAll" + entityName + "s"
	getByIdMethodName := "Get" + entityName
	postMethodName := "Post" + entityName
	putMethodName := "Put" + entityName
	deleteMethodName := "Delete" + entityName

	allMethodName := "GetAll" + entityName + "sSubEntities"
	allMethodExist := false

	specialMethods := []EntityRelationMethod{}

	//modelFile.Empty()
	////write routes in init method
	//modelFile.Comment("Routes related to " + entityName)
	//modelFile.Func().Id("init").Params().BlockFunc(func(g *Group) {
	//
	//	g.Empty()
	//	g.Comment("Standard routes")
	//	g.Qual(const_RouterPath, "Get").Call(Lit("/"+strings.ToLower(entityName)), Id(getAllMethodName))
	//	g.Qual(const_RouterPath, "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id"), Id(getByIdMethodName))
	//	g.Qual(const_RouterPath, "Post").Call(Lit("/"+strings.ToLower(entityName)), Id(postMethodName))
	//	g.Qual(const_RouterPath, "Put").Call(Lit("/"+strings.ToLower(entityName)+"/:id"), Id(putMethodName))
	//	g.Qual(const_RouterPath, "Delete").Call(Lit("/"+strings.ToLower(entityName)+"/:id"), Id(deleteMethodName))
	//
	//	if len(entityRelationsForEachEndpoint) > 0 {
	//		g.Empty()
	//		g.Comment("Sub entities routes")
	//		for _, entRel := range entityRelationsForEachEndpoint {
	//
	//			if entRel.Type == const_OneToMany {
	//				methodName := "Get" + entityName + entRel.SubEntityName + "s"
	//				specialMethods = append(specialMethods, EntityRelationMethod{methodName, entRel.Type, entRel.SubEntityName, entRel.SubEntityColName})
	//				g.Empty()
	//				g.Comment("has many")
	//				g.Qual(const_RouterPath, "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/"+strings.ToLower(entRel.SubEntityName+"s")), Id(methodName))
	//			} else if entRel.Type == const_OneToOne+const_normal || entRel.Type == const_OneToOne+const_self || entRel.Type == const_OneToOne+const_reverse {
	//				methodName := "Get" + entityName + entRel.SubEntityName
	//				specialMethods = append(specialMethods, EntityRelationMethod{methodName, entRel.Type, entRel.SubEntityName, entRel.SubEntityColName})
	//				g.Empty()
	//				g.Comment("has one")
	//				g.Qual(const_RouterPath, "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/"+strings.ToLower(entRel.SubEntityName)), Id(methodName))
	//			} else if entRel.Type == const_ManyToOne {
	//				methodName := "Get" + entityName + entRel.SubEntityName + ""
	//				specialMethods = append(specialMethods, EntityRelationMethod{methodName, entRel.Type, entRel.SubEntityName, entRel.SubEntityColName})
	//				g.Empty()
	//				g.Comment("belongs to")
	//				g.Qual(const_RouterPath, "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/"+strings.ToLower(entRel.SubEntityName)), Id(methodName))
	//			} else if entRel.Type == const_ManyToMany {
	//				methodName := "Get" + entityName + entRel.SubEntityName + "s"
	//				specialMethods = append(specialMethods, EntityRelationMethod{methodName, entRel.Type, entRel.SubEntityName, entRel.SubEntityColName})
	//				g.Empty()
	//				g.Comment("has many to many")
	//				g.Qual(const_RouterPath, "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/"+strings.ToLower(entRel.SubEntityName)), Id(methodName))
	//			}
	//
	//		}
	//	}
	//
	//	if len(entityRelationsForAllEndpoint) > 0 {
	//		allMethodExist = true
	//		g.Empty()
	//		g.Comment("extra route")
	//		g.Qual(const_RouterPath, "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/all"), Id(allMethodName))
	//	}
	//})

	createEntitiesChildSlice(modelFile, entityName, entityRelationsForAllEndpoint)

	createEntitiesGetAllMethod(modelFile, entityName, getAllMethodName)

	createEntitiesGetMethod(modelFile, entityName, getByIdMethodName)

	createEntitiesPostMethod(modelFile, entityName, postMethodName, entityFields)

	createEntitiesPutMethod(modelFile, entityName, putMethodName)

	createEntitiesDeleteMethod(modelFile, entityName, deleteMethodName)

	if len(specialMethods) > 0 {
		for _, method := range specialMethods {
			modelFile.Empty()
			modelFile.Func().Id(method.MethodName).Params(handlerRequestParams()).BlockFunc(func(g *Group) {
				g.Empty()
				g.Comment("Get the parameter id")
				g.Id("params").Op(":=").Qual(const_RouterPath, "Params").Call(Id("req"))
				g.Id("ID").Op(",").Id("_").Op(":=").Qual("strconv", "ParseUint").Call(
					Qual("", "params.ByName").Call(Lit("id")),
					Id("10"),
					Id("0"),
				)

				if method.Type == const_OneToMany || method.Type == const_OneToOne+const_normal {
					g.Id("data").Op(":= []").Id(method.SubEntityName).Id("{}")
					g.Qual(const_DatabasePath, "SQL.Find").Call(Id("&").Id("data"), Lit(" "+method.SubEntityColName+" = ?"), Id("ID"))
					g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
					g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
						Op("{").
						Id("2000").Op(",").
						Lit("Data fetched successfully").Op(",").
						Id("data").
						Op("}"))
				}

				if method.Type == const_ManyToOne || method.Type == const_OneToOne+const_reverse {
					g.Id(strings.ToLower(entityName)).Op(":=").Id(entityName).Op("{").Id("Id").Op(":").Id("uint(").Id("ID").Op(")}")

					g.Id("data").Op(":= ").Id(method.SubEntityName).Id("{}")
					g.Qual(const_DatabasePath, "SQL.Find").Call(
						Id("&").Id("data"), Lit(" id = (?)"),
						Qual(const_DatabasePath, "SQL.Select").Call(Lit(method.SubEntityColName)).Op(".").Id("First").Call(Id("&").Id(strings.ToLower(entityName))).Op(".").Id("QueryExpr").Call(),
					)
					g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
					g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
						Op("{").
						Id("2000").Op(",").
						Lit("Data fetched successfully").Op(",").
						Id("data").
						Op("}"))
				}

				if method.Type == const_OneToOne+const_self {
					g.Id("data").Op(":= ").Id(method.SubEntityName).Id("{}")
					g.Qual(const_DatabasePath, "SQL.Find").Call(Id("&").Id("data"), Lit(" "+method.SubEntityColName+" = ?"), Id("ID"))
					g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
					g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
						Op("{").
						Id("2000").Op(",").
						Lit("Data fetched successfully").Op(",").
						Id("data").
						Op("}"))
				}

				if method.Type == const_ManyToMany {

					relation := method.SubEntityName + "s"

					g.Id("data").Op(":=").Id(entityName).Id("{}")
					g.Qual(const_DatabasePath, "SQL.Find").Call(Id("&").Id("data"), Id("ID"))
					g.Qual(const_DatabasePath, "SQL.Model").Call(Id("&").Id("data")).Op(".").Id("Association").Call(Lit(relation)).
						Op(".").Id("Find").Call(Id("&").Id("data").Op(".").Id(relation))
					g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
					g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
						Op("{").
						Id("2000").Op(",").
						Lit("Data fetched successfully").Op(",").
						Id("data").
						Op("}"))
				}
			})
		}
	}

	if allMethodExist {
		createEntitiesAllChildMethod(modelFile, entityName, allMethodName, entityRelationsForAllEndpoint)
	}

	fmt.Fprintf(file, "%#v", modelFile)

	fmt.Println(entityName + " generated")
	return entityName
}

func createEntitiesChildSlice(modelFile *File, entityName string, entityRelationsForAllEndpoint []EntityRelation) {
	allChildren := []string{}
	for _, value := range entityRelationsForAllEndpoint {
		allChildren = append(allChildren, value.SubEntityName)
	}

	modelFile.Empty()
	modelFile.Comment("Child entities")
	modelFile.Var().Id(entityName + "Children").Op("=").Lit(allChildren)
}

func createEntitiesGetAllMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write getAll method
	modelFile.Comment("This method will return a list of all " + entityName + "s")
	modelFile.Func().Id(methodName).Params().Id("[]").Id(entityName).Block(
		Id("data").Op(":=").Op("[]").Id(entityName).Op("{}"),
		Qual(const_DatabasePath, "SQL.Find").Call(Id("&").Id("data")),
		Return(Id("data")),
	)
}

func createEntitiesGetMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write getOne method
	modelFile.Comment("This method will return one " + entityName + " based on id")
	modelFile.Func().Id(methodName).Params(Id("ID").Uint()).Id(entityName).Block(
		Id("data").Op(":=").Id(entityName).Op("{}"),
		Qual(const_DatabasePath, "SQL.First").Call(Id("&").Id("data"), Id("ID")),
		Return(Id("data")),
	)
}

func createEntitiesPostMethod(modelFile *File, entityName string, methodName string, entityFields []EntityField) {
	modelFile.Empty()
	//write insert method
	modelFile.Comment("This method will insert one " + entityName + " in db")
	modelFile.Func().Id(methodName).Params(Id("data").Id(entityName)).Id(entityName).Block(
		Qual(const_DatabasePath, "SQL.Create").Call(Id("&").Id("data")),
		Return(Id("data")),
	)
}

func createEntitiesPutMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write update method
	modelFile.Comment("This method will update " + entityName + " based on id")
	modelFile.Func().Id(methodName).Params(Id("newData").Id(entityName)).Id(entityName).Block(
		Id("oldData").Op(":=").Id(entityName).Id("{").Id("Id").Op(":").Id("newData").Op(".").Id("Id").Id("}"),
		Qual(const_DatabasePath, "SQL.Model").Call(Id("&oldData")).Op(".").Id("Updates").Call(Id("newData")),
		Return(Id("newData")),
	)
}

func createEntitiesDeleteMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write delete method
	modelFile.Comment("This method will delete " + entityName + " based on id")
	modelFile.Func().Id(methodName).Params(Id("ID").Uint()).Id(entityName).Block(
		Id("data").Op(":=").Id(entityName).Op("{").Id("Id").Op(":").Id("ID").Op("}"),
		Qual(const_DatabasePath, "SQL.Delete").Call(Id("&").Id("data")),
		Return(Id("data")),
	)
}

func createEntitiesAllChildMethod(modelFile *File, entityName string, allMethodName string, entityRelationsForAllEndpoint []EntityRelation) {
	modelFile.Empty()
	modelFile.Func().Id(allMethodName).Params(handlerRequestParams()).BlockFunc(func(g *Group) {
		g.Empty()
		g.Comment("Get the parameter id")
		g.Id("params").Op(":=").Qual(const_RouterPath, "Params").Call(Id("req"))
		g.Id("ID").Op(",").Id("_").Op(":=").Qual("strconv", "ParseUint").Call(
			Qual("", "params.ByName").Call(Lit("id")),
			Id("10"),
			Id("0"),
		)
		g.Id("data").Op(":=").Id(entityName).Op("{").Id("Id").Op(":").Id("uint(ID)").Op("}")
		g.Empty()
		g.Var().Id("relations ").Op("[").Id(strconv.Itoa(len(entityRelationsForAllEndpoint))).Op("]").Id("string")
		g.Id("children").Op(":=").Qual("", "req.URL.Query().Get").Call(Lit("child"))
		g.If(Id("children").Op("!= \"\"")).
			Block(
			Var().Id("neededChildren ").Op("[]").Id("string"),

			For(Id("_,child").Op(":=").Id("range").Id(entityName + "Children")).
				Block(
				If(Qual("", "isValueInList").
					Call(
					Id("child"),
					Qual("strings", "Split").
						Call(
						Id("children"), Id("sep"),
					),
				).
					Block(
					Id("neededChildren").Op("=").Qual("", "append").Call(Id("neededChildren"), Id("child")),
				),
				), ),

			Empty(),

			For(Id("i").Op(":=").Id("range").Id("neededChildren")).
				Block(
				Id("relations").Op("[").Id("i").Op("]").Op("=").Id("neededChildren").Op("[").Id("i").Op("]"),
			),
		).Else().
			Block(
			For(Id("i").Op(":=").Id("range").Id(entityName + "Children")).
				Block(
				Id("relations").Op("[").Id("i").Op("]").Op("=").Id(entityName + "Children").Op("[").Id("i").Op("]"),
			),
		)
		g.If(Qual("", "len").Call(Id("relations")).Op(">0")).BlockFunc(func(g *Group) {

			var buffer bytes.Buffer
			buffer.WriteString("SQL.")
			for i := range entityRelationsForAllEndpoint {
				buffer.WriteString("Preload(relations[" + strconv.Itoa(i) + "]).")
			}
			buffer.WriteString("First")
			g.Qual(const_DatabasePath, buffer.String()).Call(Op("&").Id("data"))
		})
		g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
		g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
			Op("{").
			Id("2000").Op(",").
			Lit("Data fetched successfully").Op(",").
			Id("data").
			Op("}"))
	})
}

func mapColumnTypes(col Column, g *Group) EntityField {

	entityField := EntityField{}
	entityField.FieldName = col.Name

	if col.ColumnType.Type == "int" {
		entityField.FieldType = "uint"
		finalId := snakeCaseToCamelCase(col.Name) + " uint" + " `gorm:\"column:" + col.Name + "\" json:\"" + col.Name + ",omitempty\"`"
		g.Id(finalId)
	} else if col.ColumnType.Type == "varchar" {
		entityField.FieldType = "string"
		finalId := snakeCaseToCamelCase(col.Name) + " string" + " `gorm:\"column:" + col.Name + "\" json:\"" + col.Name + ",omitempty\"`"
		g.Id(finalId)
	} else {
		entityField.FieldType = "string"
		g.Id(snakeCaseToCamelCase(col.Name)).String() //default string
	}
	return entityField
}

//helper methods
func snakeCaseToCamelCase(inputUnderScoreStr string) (camelCase string) {
	//snake_case to camelCase

	isToUpper := false

	for k, v := range inputUnderScoreStr {
		if k == 0 {
			camelCase = strings.ToUpper(string(inputUnderScoreStr[0]))
		} else {
			if isToUpper {
				camelCase += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				if v == '_' {
					isToUpper = true
				} else {
					camelCase += string(v)
				}
			}
		}
	}
	return

}

func handlerRequestParams() (Code, Code) {
	return Id("w").Qual("net/http", "ResponseWriter"), Id("req").Op("*").Qual("net/http", "Request")
}

func setJsonHeader() Code {
	return Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
}

func sendResponse(statusCode uint, statusMsg string, data interface{}) Code {
	return Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
		Op("{").
		Lit(statusCode).Op(",").
		Lit(statusMsg).Op(",").
		Lit(data).
		Op("}"))
}
