package connectionhub

type MesageTarget string

const (
	User         MesageTarget = "user"
	Organization MesageTarget = "organization"
	Role         MesageTarget = "role"
)

func (t MesageTarget) String() string {
	switch t {
	case User:
		return "user"
	case Organization:
		return "organization"
	case Role:
		return "role"
	}
	return "unknown"
}
