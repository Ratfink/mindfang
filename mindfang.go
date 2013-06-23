package main

import (
    "log"
    "os"
	"io"
	"math"
    "strings"
    "bufio"
	"image/color"
    "github.com/ziutek/serial"
	"github.com/Ratfink/gopherbone/gpio"
	"github.com/Ratfink/gopherbone/ssd1306"
)

const (
	scrollback = 256
)

func ButtonGet(button chan byte) (err error) {
	evt_file, err := os.OpenFile("/dev/input/event1", os.O_RDWR, 0666)
	if err != nil {
		return
	}
	defer evt_file.Close()
	evt_buf := make([]byte, 16)

	for {
		_, err = evt_file.Read(evt_buf)
		if err != nil {
			return
		}
		_, err = evt_file.Read(nil)
		if err != nil {
			return
		}
		// Don't send up events. This behaviour may change in the future if I
		// try to implement key repeating. Maybe Linux itself supports that,
		// though. TODO: research this topic more.
		if evt_buf[12] != 0 {
			button <- evt_buf[10]
		}
	}
	return
}

func ReadInput(f *os.File, msg chan string, s *serial.Serial, eof chan error) {
    rpreader := bufio.NewReader(s)
//    rpwriter := bufio.NewWriter(s)
	var str string
	var err error

	for {
		str, err = rpreader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				eof <- err
			}
			return
		}
		msg <- str
	}
}

func main() {
	button := make(chan byte, 16)
	msg := make(chan string, 16)
	eof := make(chan error, 0)
	eofSeen := false
	buf := make([]string, scrollback)
	pos := 0
	var nlines float64 = 0
	var str string
	var btn byte
	disp, err := ssd1306.New(gpio.P8[3], ssd1306.IFACE_I2C, 0x3c, 1, 128, 64)
	defer disp.Close()
	if err != nil {
		log.Fatal(err)
	}
	disp.Setup()

    var s *serial.Serial

    s, err = serial.Open(os.Args[len(os.Args)-1])
    if err != nil {
        log.Fatal(err)
    }
    defer s.Close()
    s.SetSpeed(115200)
    s.SetCanon(true)

//    cts := make(chan bool)

	go ButtonGet(button)
	go ReadInput(os.Stdin, msg, s, eof)
	for !eofSeen {
		disp.Clear(color.Black)
		for i := 0; i < 8; i++ {
			disp.String(0, 63 - 8*i, color.White, buf[pos+i])
		}
        disp.Rectangle(126, int(64-64*(float64(pos)+8)/math.Max(nlines, 8)),
                       127, int(63-64*float64(pos)/math.Max(nlines, 8)), color.White)
		disp.Draw()
		select {
		case str = <-msg:
			for long := true; long; {
				for i := int(nlines); i >= 1; i-- {
					buf[i] = buf[i-1]
				}
				str = strings.TrimSpace(str)
				buf[0] = str[:int(math.Min(float64(len(str)), 21))]
				if nlines < scrollback {
					nlines++
				}
				if len(str) <= 21 {
					long = false
				} else {
					str = str[21:]
				}
			}
		case btn = <-button:
			switch btn {
			case 1:
				if pos < int(nlines) - 8 {
					pos++
				}
			case 2:
				if pos > 0 {
					pos--
				}
			}
		case <-eof:
			eofSeen = true
		}
	}
}
