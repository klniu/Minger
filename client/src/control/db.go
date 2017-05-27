package control

import (
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "github.com/jinzhu/gorm"
    "fmt"
    "github.com/qor/validations"
    "strings"
    "reflect"
    . "Minger/client/src/model"
)

type DB struct {
    *gorm.DB
}

func Connect(database string) (*DB, error) {
    db := new(DB)
    var err error
    if db.DB, err = gorm.Open("sqlite3", database); err != nil {
        return db, fmt.Errorf("打开数据库失败：%s", err.Error())
    }
    // debug
    // db = &DB{db.Debug()}
    if err = db.runMigrate(); err != nil {
        return db, err
    }
    validations.RegisterCallbacks(db.DB)
    return db, err
}

func (d *DB) runMigrate() error {
    if err := d.AutoMigrate(PlayList{}, Audio{}).Error; err != nil {
        return fmt.Errorf("创建数据库失败：%s", err.Error())
    }
    var option Option
    var err error
    if option, err = d.ReadOption("IsNewDatabase"); err != nil && err != gorm.ErrRecordNotFound {
        return err
    }
    if option.Value == "1" {
        return nil
    }
    // first initialize
    // create admin
    if err := d.createAdminUser(); err != nil {
        return fmt.Errorf("创建管理员失败：%s", err.Error())
    }
    if err := d.Create(&Option{"IsNewDatabase", "1"}).Error; err != nil {
        return fmt.Errorf("数据库初始化失败：%s", err.Error())
    }
    return nil
}

// exec execute insert, update or delete for slices. operation can be "insert, update, or delete".
func (d *DB) exec(operation string, args ...interface{}) error {
    var err error
    // begin
    tx := d.Begin()
    // here we have not check the rationality of the operating object, such as material and its id.
    operate := func(obj interface{}) error {
        switch operation {
        case "insert":
            err = tx.Create(obj).Error
        case "update":
            err = tx.Save(obj).Error
        case "delete":
            err = tx.Delete(obj).Error
        }
        if err != nil {
            // rollback
            if err.Error() == "FOREIGN KEY constraint failed" {
                return fmt.Errorf("无关联数据")
            } else if strings.Contains(err.Error(), "UNIQUE constraint failed") {
                return fmt.Errorf("重复的数据")
            } else {
                return err
            }
        }
        return nil
    }
    var (
        valType reflect.Kind
        val     reflect.Value
    )
    for i := range args {
        valType = reflect.TypeOf(args[i]).Kind()
        val = reflect.ValueOf(args[i])
        switch valType {
        case reflect.Slice, reflect.Array:
            for i := 0; i < val.Len(); i++ {
                err = operate(val.Index(i).Addr().Interface())
                if err != nil {
                    tx.Rollback()
                    return err
                }
            }
        case reflect.Ptr:
            err = operate(val.Interface())
            if err != nil {
                tx.Rollback()
                return err
            }
        default:
            tx.Rollback()
            return fmt.Errorf("Unsupported type")
        }

    }
    return tx.Commit().Error
}

// Insert insert objects (slice) into tables
func (d *DB) Insert(args ...interface{}) error {
    return d.exec("insert", args...)
}

// Update update objects in tables
func (d *DB) Update(args ...interface{}) error {
    return d.exec("update", args...)
}

// Delete delete objects into tables
func (d *DB) Delete(args ...interface{}) error {
    return d.exec("delete", args...)
}

