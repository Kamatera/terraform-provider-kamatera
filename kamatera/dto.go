package kamatera

type listServersPostValues struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type createServerPostValues struct {
	Name             string `json:"name"`
	Password         string `json:"password"`
	PasswordValidate string `json:"passwordValidate"`
	SSHKey           string `json:"ssh-key"`
	Datacenter       string `json:"datacenter"`
	Image            string `json:"image"`
	CPU              string `json:"cpu"`
	RAM              float64 `json:"ram"`
	Disk             string `json:"disk"`
	DailyBackup      string `json:"dailybackup"`
	Managed          string `json:"managed"`
	Network          string `json:"network"`
	Quantity         string `json:"quantity"`
	BillingCycle     string `json:"billingcycle"`
	MonthlyPackage   string `json:"monthlypackage"`
	PowerOn          string `json:"poweronaftercreate"`
}

type powerOperationServerPostValues struct {
	ID    string `json:"id"`
	Force bool   `json:"force"`
}

type configureServerPostValues struct {
	ID string `json:"id"`
	CPU string `json:"cpu"`
	RAM float64 `json:"ram"`
	DailyBackup string `json:"dailybackup"`
	Managed string `json:"managed"`
	BillingCycle string `json:"billingcycle"`
	MonthlyPackage string `json:"monthlypackage"`
}

type changePasswordServerPostValues struct {
	ID string `json:"id"`
	Password string `json:"password"`
}

type renameServerPostValues struct {
	ID string `json:"id"`
	NewName string `json:"new-name"`
}
