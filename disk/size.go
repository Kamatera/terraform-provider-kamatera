package disk

type Size int

//go:generate enumer -type=Size -trimprefix=Size
const (
	Size10GB Size = iota
	Size20GB
	Size30GB
	Size40GB
	Size50GB
	Size60GB
	Size80GB
	Size100GB
	Size150GB
	Size200GB
	Size250GB
	Size300GB
	Size350GB
	Size400GB
	Size450GB
	Size500GB
	Size600GB
	Size700GB
	Size800GB
	Size900GB
	Size1TB
	// Size15TB
	Size2TB
	Size3TB
	Size4TB
)
