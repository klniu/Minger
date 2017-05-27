package model

import (
    "errors"
    "fmt"
    "strings"
    "time"

    "golang.org/x/crypto/bcrypt"

    "github.com/jinzhu/gorm"
    "github.com/qor/validations"
)

// 账户
type User struct {
    ID             int
    UserName       string    `gorm:"not null;unique;size:50" valid:"required"`       // 账户ID
    Password       string    `gorm:"not null;size:100"`                       // 账户密码
    Email          string    `gorm:"not null;unique;size:50" valid:"email""` // email
    Sex            int                                                       // 1:男， 2：女
    PNumber        int64                                                     // 手机号
    CreateAt       time.Time `gorm:"not null"`                               // 创建时间
    LoginAt        time.Time                                                 // 登录时间
    LogoutAt       time.Time                                                 // 登出时间
    LoginCount     int64                                                     // 登录次数
    LoginFailCount int                                                       // 登录失败次数
    Role           int                                                       // 角色
}

type Role int

const (
    Admin   int = iota // 管理员，可以添加用户，处理任何内容
    Editor             // 添加修改任何内容
    Visitor            // 仅查看文章
)

func (u *User) Validate(db *gorm.DB) {
    if strings.TrimSpace(u.UserName) == "" {
        db.AddError(validations.NewError(u, "UserName", "用户名不能为空"))
    }
    if len(u.UserName) > 50 {
        db.AddError(validations.NewError(u, "UserName", "用户名过长"))
    }
    if len(u.Password) > 100 {
        db.AddError(validations.NewError(u, "UserName", "密码过长"))
    }
    var isok bool
    for _, r := range []int{Admin, Editor, Visitor} {
        if u.Role == r {
            isok = true
            break
        }
    }
    if !isok {
        db.AddError(validations.NewError(u, "Role", "角色错误"))
    }
}

func (u *User) cryptPass() error {
    cryptPass, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(cryptPass)
    return nil
}

func (d *DB) createAdminUser() error {
    // 新建一个超级管理员用户
    user := User{UserName: "admin", Password: "f865b53623b121fd34ee5426c792e5c33af8c227",
        Email:             "example@example.org", Role: Admin, CreateAt: time.Now()}
    if err := user.cryptPass(); err != nil {
        return fmt.Errorf("超级管理员密码加密失败: %s", err.Error())
    }
    var cnt int
    if err := d.Model(User{}).Where("user_name=?", user.UserName).Count(&cnt).Error; err != nil {
        return fmt.Errorf("查询管理员帐号失败: %s", err.Error())
    }
    if cnt == 0 {
        if err := d.Create(&user).Error; err != nil {
            return fmt.Errorf("创建超级管理员失败: %s", err.Error())
        }
    }
    return nil
}

func (d *DB) CreateUser(user User) error {
    if err := user.cryptPass(); err != nil {
        return errors.New("用户密码加密失败")
    }
    if err := d.Create(&user).Error; err != nil {
        return fmt.Errorf("创建用户失败")
    }
    return nil
}

func (d *DB) Login(username, pass string) (user User, err error) {
    if strings.TrimSpace(username) == "" {
        return user, errors.New("用户名不能为空")
    }
    err = d.Where("user_name=?", username).First(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return user, errors.New("用户不存在")
        }
        return user, errors.New("查找用户失败")
    }
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
    if err != nil {
        return user, errors.New("密码错误")
    }

    return
}

func (d *DB) ChangePassword(userID int, oriPass, newPass string) error {
    var user User
    err := d.Where("id=?", userID).First(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return errors.New("用户不存在")
        }
        return errors.New("查找用户失败")
    }
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oriPass))
    if err != nil {
        return errors.New("原始密码错误")
    }
    user.Password = newPass
    if err = user.cryptPass(); err != nil {
        return errors.New("用户密码加密失败")
    }
    if err = d.Save(&user).Error; err != nil {
        return fmt.Errorf("密码更新失败")
    }
    return nil
}
