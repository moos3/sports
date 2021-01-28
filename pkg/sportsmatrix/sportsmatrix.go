package sportsmatrix

import (
	"context"
	"fmt"
	"image"
	_ "image/png"
	"time"

	rgb "github.com/robbydyer/sports/pkg/rgbmatrix-rpi"

	"github.com/robbydyer/sports/pkg/board"
)

type SportsMatrix struct {
	cfg    *Config
	matrix rgb.Matrix
	boards []board.Board
	done   chan bool
}

type Config struct {
	RotationDelay  string
	EnableNHL      bool `json:"enableNHL,omitempty"`
	HardwareConfig *rgb.HardwareConfig
}

func (c *Config) Defaults() {
	if c.HardwareConfig == nil {
		c.HardwareConfig = &rgb.DefaultConfig
	}

	if c.HardwareConfig.Rows == 0 {
		c.HardwareConfig.Rows = 32
	}
	if c.HardwareConfig.Cols == 32 || c.HardwareConfig.Cols == 0 {
		// We don't support 32x32 matrix
		c.HardwareConfig.Cols = 64
	}
	// The defaults do 100, but that's too much
	if c.HardwareConfig.Brightness == 0 || c.HardwareConfig.Brightness == 100 {
		c.HardwareConfig.Brightness = 60
	}
	if c.RotationDelay == "" {
		c.RotationDelay = "20s"
	}
	if c.HardwareConfig.HardwareMapping == "" {
		c.HardwareConfig.HardwareMapping = "adafruit-hat-pwm"
	}
	if c.HardwareConfig.ChainLength == 0 {
		c.HardwareConfig.ChainLength = 1
	}
	if c.HardwareConfig.Parallel == 0 {
		c.HardwareConfig.Parallel = 1
	}
	if c.HardwareConfig.PWMBits == 0 {
		c.HardwareConfig.PWMBits = 11
	}
	if c.HardwareConfig.PWMLSBNanoseconds == 0 {
		c.HardwareConfig.PWMLSBNanoseconds = 130
	}
}

func (c *Config) rotationDelay() time.Duration {
	d, err := time.ParseDuration(c.RotationDelay)
	if err != nil {
		fmt.Printf("could not parse duration '%s', defaulting to 20 sec", c.RotationDelay)
		return 20 * time.Second
	}
	return d
}

func New(ctx context.Context, cfg *Config, boards ...board.Board) (*SportsMatrix, error) {
	cfg.Defaults()

	s := &SportsMatrix{
		boards: boards,
		cfg:    cfg,
		done:   make(chan bool, 1),
	}

	var err error

	fmt.Printf("Initializing matrix %dx%d\nBrightness:%d\nMapping:%s\n",
		s.cfg.HardwareConfig.Cols,
		s.cfg.HardwareConfig.Rows,
		s.cfg.HardwareConfig.Brightness,
		s.cfg.HardwareConfig.HardwareMapping,
	)

	rt := &rgb.DefaultRuntimeOptions
	s.matrix, err = rgb.NewRGBLedMatrix(s.cfg.HardwareConfig, rt)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// MatrixBounds returns an image.Rectangle of the matrix bounds
func (s *SportsMatrix) MatrixBounds() image.Rectangle {
	w, h := s.matrix.Geometry()
	return image.Rect(0, 0, w-1, h-1)
}

func (s *SportsMatrix) Done() chan bool {
	return s.done
}

func (s *SportsMatrix) Serve(ctx context.Context) error {
	if len(s.boards) < 1 {
		return fmt.Errorf("no boards configured")
	}
	fmt.Println("Serving boards...")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Got context cancel, cleaning up boards")
			go func() {
				for _, b := range s.boards {
					b.Cleanup()
				}
			}()
			s.done <- true
			return nil
		default:
		}
	INNER:
		for _, b := range s.boards {
			if s.anyPriorities() && !b.HasPriority() {
				continue
			}
			if b.HasPriority() {
				fmt.Printf("Rendering board '%s' as priority\n", b.Name())
				err := b.Render(ctx, s.matrix)
				if err != nil {
					fmt.Printf("Error: %s", err.Error())
				}
				break INNER
			}
			fmt.Printf("Rendering board '%s'\n", b.Name())
			err := b.Render(ctx, s.matrix)
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
			}
			b.Cleanup()
		}
	}
}

func (s *SportsMatrix) anyPriorities() bool {
	for _, b := range s.boards {
		if b.HasPriority() {
			return true
		}
	}

	return false
}

func (s *SportsMatrix) Close() {
	if len(s.boards) > 1 {
		fmt.Println("Waiting for boards to clean up")
		<-s.done
	}
	if s.matrix != nil {
		fmt.Println("Closing matrix")
		_ = s.matrix.Close()
	}
}