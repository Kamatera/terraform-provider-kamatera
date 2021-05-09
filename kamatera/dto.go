package kamatera

type listServersPostValues struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type createServerPostValues struct {
	Name             string `json:"name"`
	Password         string `json:"password"`
	PasswordValidate string `json:"passwordValidate"`
	SshKey           string `json:"ssh-key"`
	Datacenter       string `json:"datacenter"`
	Image            string `json:"image"`
	Cpu              string `json:"cpu"`
	Ram              int64  `json:"ram"`
	Disk             string `json:"disk"`
	DailyBackup      string `json:"dailybackup"`
	Managed          string `json:"managed"`
	Network          string `json:"network"`
	Quantity         string `json:"quantity"`
	BillingCycle     string `json:"billingcycle"`
	MonthlyPackage   string `json:"monthlypackage"`
	PowerOn          string `json:"poweronaftercreate"`
}
