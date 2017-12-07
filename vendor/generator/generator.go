package generator

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
	ParentEntityID    int `sql:"type:int(100)"`
	ParentEntityColID int `sql:"type:int(100)"`
	ChildEntityID     int `sql:"type:int(100)"`
	ChildEntityColID  int `sql:"type:int(100)"`
	InterEntityID     int `sql:"type:int(100)"`
	RelationTypeID    int `sql:"type:int(10)"`

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
