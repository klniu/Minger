package model

type Option struct {
    Name  string `gorm:"type:varchar(100);primary_key"`
    Value string `gorm:"type:varchar(100);"`
}

