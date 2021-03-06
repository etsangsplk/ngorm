package ngorm

import (
	"fmt"
	"log"
	"sort"

	"github.com/ngorm/ngorm/model"

	"strings"
)

func ExampleOpen() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	fmt.Println(db.Dialect().GetName())

	//Output:ql-mem
}
func ExampleDB_CreateSQL() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	type Bar struct {
		ID  int64
		Say string
	}

	b := Bar{Say: "hello"}
	sql, err := db.CreateSQL(&b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sql.Q)
	fmt.Printf("$1=%v", sql.Args[0])

	//Output:
	//BEGIN TRANSACTION;
	//	INSERT INTO bars (say) VALUES ($1);
	//COMMIT;
	//$1=hello
}

func ExampleDB_CreateSQL_extraOptions() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	type Bar struct {
		ID  int64
		Say string
	}

	b := Bar{Say: "hello"}
	sql, err := db.Set(model.InsertOptions, "ON CONFLICT").
		CreateSQL(&b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sql.Q)
	fmt.Printf("$1=%v", sql.Args[0])

	//Output:
	//BEGIN TRANSACTION;
	//	INSERT INTO bars (say) VALUES ($1) ON CONFLICT;
	//COMMIT;
	//$1=hello
}

func ExampleDB_AutomigrateSQL() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	type Bar struct {
		ID  int64
		Say string
	}

	type Bun struct {
		ID   int64
		Dead bool
	}

	sql, err := db.AutomigrateSQL(&Bar{}, &Bun{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sql.Q)

	//Output:
	//BEGIN TRANSACTION;
	//	CREATE TABLE bars (id int64,say string ) ;
	//	CREATE TABLE buns (id int64,dead bool ) ;
	//COMMIT;
}

func ExampleDB_Automigrate() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	type Bar struct {
		ID  int64
		Say string
	}

	type Bun struct {
		ID   int64
		Dead bool
	}

	_, err = db.Automigrate(&Bar{}, &Bun{})
	if err != nil {
		log.Fatal(err)
	}
	var names []string
	qdb := db.SQLCommon()
	rows, err := qdb.Query("select Name  from __Table ")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var n string
		err = rows.Scan(&n)
		if err != nil {
			log.Fatal(err)
			fmt.Println(err)
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	sort.Strings(names)
	for _, v := range names {
		fmt.Println(v)
	}

	//Output:
	//bars
	//buns

}

func Example_belongsTo() {
	//One to many relationship
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	// Here one user can only have one profile. But one profile can belong
	// to multiple users.
	type Profile struct {
		model.Model
		Name string
	}
	type User struct {
		model.Model
		Profile   Profile
		ProfileID int
	}
	_, err = db.Automigrate(&User{}, &Profile{})
	if err != nil {
		log.Fatal(err)
	}
	u := User{
		Profile: Profile{Name: "gernest"},
	}

	// Creating the user with the Profile. The relation will
	// automatically be created and the user.ProfileID will be set to
	// the ID of hte created profile.
	err = db.Create(&u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(u.ProfileID == int(u.Profile.ID) && u.ProfileID != 0)

	//Output:
	//true

}

func Example_migration() {
	type Profile struct {
		model.Model
		Name string
	}
	type User struct {
		model.Model
		Profile   Profile
		ProfileID int64
	}

	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	// you can inspect expected generated query
	s, err := db.AutomigrateSQL(&User{}, &Profile{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s.Q)

	// Or you can execute migrations like so
	_, err = db.Begin().Automigrate(&User{}, &Profile{})
	if err != nil {
		log.Fatal(err)
	}
	//Output:
	// BEGIN TRANSACTION;
	// 	CREATE TABLE users (id int64,created_at time,updated_at time,deleted_at time,profile_id int64 ) ;
	// 	CREATE INDEX idx_users_deleted_at ON users(deleted_at);
	// 	CREATE TABLE profiles (id int64,created_at time,updated_at time,deleted_at time,name string ) ;
	// 	CREATE INDEX idx_profiles_deleted_at ON profiles(deleted_at);
	// COMMIT;
}
func ExampleDB_SaveSQL() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	type Foo struct {
		ID    int64
		Stuff string
	}

	f := Foo{ID: 10, Stuff: "twenty"}
	sql, err := db.SaveSQL(&f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sql.Q)
	fmt.Printf("$1=%v\n", sql.Args[0])
	fmt.Printf("$2=%v", sql.Args[1])

	//Output:
	//BEGIN TRANSACTION;
	//	UPDATE foos SET stuff = $1  WHERE id = $2;
	//COMMIT;
	//$1=twenty
	//$2=10
}

func ExampleDB_UpdateSQL() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	type Foo struct {
		ID    int64
		Stuff string
	}

	f := Foo{ID: 10, Stuff: "twenty"}
	sql, err := db.Model(&f).UpdateSQL("stuff", "hello")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sql.Q)
	fmt.Printf("$1=%v\n", sql.Args[0])
	fmt.Printf("$2=%v", sql.Args[1])

	//Output:
	//BEGIN TRANSACTION;
	//	UPDATE foos SET stuff = $1  WHERE id = $2;
	//COMMIT;
	//$1=hello
	//$2=10
}

func ExampleDB_Find() {
	type User struct {
		ID   int64
		Name string
	}

	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	_, err = db.Automigrate(&User{})
	if err != nil {
		log.Fatal(err)
	}
	v := []string{"gernest", "kemi", "helen"}
	for _, n := range v {
		err = db.Begin().Save(&User{Name: n})
		if err != nil {
			log.Fatal(err)
		}
	}

	users := []User{}
	err = db.Begin().Find(&users)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range users {
		fmt.Println(u.Name)
	}

	//Output:
	// helen
	// kemi
	// gernest
}

func ExampleDB_CreateTable() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	type User struct {
		ID       int64
		Name     string
		Password string
		Email    string
	}
	_, err = db.CreateTable(&User{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(db.Dialect().HasTable("users"))
	//Output:
	//true

}

func ExampleDB_CreateTableSQL() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	type User struct {
		ID       int64
		Name     string
		Password string
		Email    string
	}
	sql, err := db.CreateTableSQL(&User{})
	if err != nil {
		log.Fatal(err)
	}
	e := `
BEGIN TRANSACTION; 
	CREATE TABLE users (id int64,name string,password string,email string ) ;
COMMIT;
`
	sql.Q = strings.TrimSpace(sql.Q)
	e = strings.TrimSpace(e)
	fmt.Println(sql.Q == e)
	//Output:
	//true
}

func ExampleDB_Count() {
	db, err := Open("ql-mem", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	type Foo struct {
		ID    int64
		Stuff string
	}
	_, err = db.Automigrate(&Foo{})
	if err != nil {
		log.Fatal(err)
	}

	sample := []string{"a", "b", "c", "d"}
	for _, v := range sample {
		err := db.Create(&Foo{Stuff: v})
		if err != nil {
			log.Fatal(err)
		}
	}
	var stuffs int64
	err = db.Model(&Foo{}).Count(&stuffs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stuffs)

	//Output:
	//4
}
