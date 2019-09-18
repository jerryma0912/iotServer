package archives

type GhostArchive struct{
	Equipment  *Equip
	Connection *ConnArchive
	AppName string
}

func NewGhostArchive(e *Equip, appName string) *GhostArchive {

	return &GhostArchive{
		Equipment: e,
		AppName:appName,
	}
}

func (g *GhostArchive) SetConnection(c *ConnArchive) {
	g.Connection = c
}

func (g *GhostArchive) ClearConnection(){
	g.Connection = nil
}