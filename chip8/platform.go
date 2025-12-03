package chip8

type Platform string

const (
	PlatformChip8   Platform = "ch8"
	PlatformSChip11 Platform = "sc"
	PlatformXOChip  Platform = "xo"
)

var ConfByPlatform = map[Platform]PlatformConf{
	PlatformChip8:   {Quirks: QuirksChip8, TickRate: 15},
	PlatformSChip11: {Quirks: QuirksSChip11, TickRate: 30},
	PlatformXOChip:  {Quirks: QuirksXOChip, TickRate: 100, AudioMode: AudioXOChip},
}
var PlatformByExt = map[string]Platform{
	".ch8": PlatformChip8,
	".sc":  PlatformSChip11,
	".sc8": PlatformSChip11,
	".xo":  PlatformXOChip,
	".xo8": PlatformXOChip,
	".8o":  PlatformXOChip,
}

type PlatformConf struct {
	Quirks    Quirks
	TickRate  int
	AudioMode AudioMode
}

func (c *PlatformConf) CPUHz() float64 {
	return float64(c.TickRate) * 60.0
}
