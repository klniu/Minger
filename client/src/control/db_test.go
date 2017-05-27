package control

import (
    . "gopkg.in/check.v1"
    . "Minger/client/src/model"
    . "Minger/common"
)

type ModelSuite struct{}

var _ = Suite(&ModelSuite{})

func emptyDB() *DB {
    // CREATE USER 'klniu'@'localhost' IDENTIFIED BY 'klniu';
    // CREATE DATABASE lawindexer;
    // GRANT ALL ON lawindexer.* TO 'klniu'@'localhost';
    d, err := Connect("robot")
    CheckErr(err)
    //d = &DB{d.Debug()}
    //var models []string
    if d.HasTable(&User{}) {
        // clear tables
        for _, modal := range []interface{}{Option{}, User{}} {
            d.Exec("TRUNCATE TABLE " + d.NewScope(modal).TableName())
            //models = append(models, d.NewScope(modal).TableName())
        }
    }
    CheckErr(d.runMigrate())
    return d
}

