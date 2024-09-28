package globalSet

var (
	MapWidth  = 200
	MapHeight = 200
)

func ToPosition(x, y int) int {
	return x + y*MapHeight
}
