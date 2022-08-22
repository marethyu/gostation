# GoStation

WIP PSX emulator in Go

![](startup.png)

TODO:
- Figure out how to load game rom
  * also find how out to get out of a infinite loop after boot logo get displayed
- CDROM
- timers
- timing
  * instruction caching
  * Using [this](https://github.com/JaCzekanski/ps1-tests/blob/master/cpu/access-time/psx.log) as reference, implement bus waitstates
- better DMA behaviour
- controller input
- interrupt handling
- start implementing GTE
- more GPU commands and other obscure GPU stuff
- optimize software renderer
  * use fixed point arithmetic
  * run renderer in a different thread?
  * many other things
- boot Crash Bandicoot
- web server for debugging
- wasm port
- savestates
