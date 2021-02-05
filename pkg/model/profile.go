package model

// ключ - MD5 от Uid пользователя (он постоянный)
// нужно, чтобы иметь возможность одному пользователю с разных браузеров иметь доступ к одному профилю
type Profile struct {
	Data map[string]*ProfileData
}

type ProfileData struct {
	Hash       		string
	Email       	string
	Uid         	string
	First_name  	string
	Last_name   	string
	Photo       	string
	Age       		string
	City        	string
	Country     	string
	Status 			string 	// - src поля Status в профиле (иногда необходимо для доп.фильтрации)
	Raw	       		[]Data	// объект пользователя (нужен при сборки проекта для данного юзера при добавлении прав на базу)
	Tables      	[]Data
	Roles       	[]Data
	Profiles       	[]Data
	Homepage		string
	Maket			string
	Groups			string
	GroupsName		string
	UpdateFlag 		bool
	UpdateData 		[]Data
	ButtonsNavTop	[]Data
	CurrentRole 	Data
	CurrentProfile 	Data
	Navigator   	[]*Items
	CountLicense	int
	BaseMode		map[string]string
}

type Items struct {
	Title  			string   	`json:"title"`
	ExtentedLink 	string 		`json:"extentedLink"`
	Uid    			string   	`json:"uid"`
	Source 			string   	`json:"source"`
	Icon   			string   	`json:"icon"`
	Leader 			string   	`json:"leader"`
	Order  			string   	`json:"order"`
	Type   			string   	`json:"type"`
	Preview			string   	`json:"preview"`
	Url    			string   	`json:"url"`
	Sub    			[]string 	`json:"sub"`
	Incl   			[]*Items 	`json:"incl"`
	Class 			string 		`json:"class"`
	FinderMode 		string 		`json:"finder_mode"`
}
