package models

type Rig struct {
	ID          int    `gorm:"primaryKey,autoIncrement" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	Description string `gorm:"column:description" json:"description"`
	Host        string `gorm:"column:host" json:"host"`
}
