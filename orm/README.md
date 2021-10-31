## devtools.ORM

devtools.ORM 是一款数据库组件, 基于 `gorm`, 下面讲下`MySQL`数据库的操作方式.

#### 连接数据库

```golang
db := orm.GetDB(
    orm.WithIp("127.0.0.1"),                // ip 
	orm.WithPort(3306),                     // port
    orm.WithUserName("rdb"),                // username
    orm.WithPassword("Root@localhost1"),    // password
    orm.WithDatabase("test"),               // database name
)
```

#### 查找数据

```golang
type Test struct {
	Id int64 `json:"id"`
	Name string `json:"name"`
}

db := orm.GetDB(
    orm.WithIp("127.0.0.1"),                // ip 
    orm.WithPort(3306),                     // port
    orm.WithUserName("rdb"),                // username
    orm.WithPassword("Root@localhost1"),    // password
    orm.WithDatabase("test"),               // database name
)

test := Test{}
err := db.Table("test").Where("id=?", 1).Find(&test).Error
if err != nil {
	panic(err)
}
fmt.Printf("%v", test)
```

#### 帮助文档

```shell
https://gorm.io/docs
```