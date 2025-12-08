package chip8

type Platform string

const (
	PlatformChip8   Platform = "ch8"
	PlatformSChip11 Platform = "sc"
	PlatformXOChip  Platform = "xo"
)

var (
	DefaultConf    = ConfByPlatform[PlatformSChip11]
	ConfByPlatform = map[Platform]PlatformConf{
		PlatformChip8:   {Quirks: QuirksChip8, Tickrate: 15},
		PlatformSChip11: {Quirks: QuirksSChip11, Tickrate: 30},
		PlatformXOChip:  {Quirks: QuirksXOChip, Tickrate: 100, AudioMode: AudioXOChip},
	}
	PlatformByExt = map[string]Platform{
		".ch":  PlatformChip8,
		".ch8": PlatformChip8,
		".sc":  PlatformSChip11,
		".sc8": PlatformSChip11,
		".xo":  PlatformXOChip,
		".xo8": PlatformXOChip,
	}
)

type PlatformConf struct {
	Quirks    Quirks
	Tickrate  int
	AudioMode AudioMode
}

func (c *PlatformConf) CPUHz() float64 {
	return float64(c.Tickrate) * 60.0
}
