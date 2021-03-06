# gotermtris

**GoTermTris** is a small weekend project to create a simple terminal-version 
of the classic game Tetris using Go.

Somewhat inspired by [javidx9](https://www.youtube.com/watch?v=8OK8_tHeCIA) 
little tutorial project doing the same in C++. And yes, I did peek and in 
places wholesale borrowed game-logic from his source, found 
[here](https://github.com/OneLoneCoder/videos/blob/master/OneLoneCoder_Tetris.cpp).

A few bells and whistles have been added such as a start-screen with top-5
list and having said list be saved on disk.

<img src="https://i.imgur.com/mk0vDkN.png" alt="From Terminator on Ubuntu 18.04" width="20%">

## Build

Contains the following non-standard library:

- github.com/gizak/termui/v3

**Contact:**

location   | name/handle |
-----------|-------------|
github:    | rDybing     |
twitter:   | @DybingRoy  |
Linked In: | Roy Dybing  |

---

## Releases

- Version format: [major release].[new feature(s)].[bugfix patch-version]

#### v.1.0.5: 19th of September 2019

- Improved name-input routine on new entry onto hi-Score list, including 
ability to backspace.

#### v.1.0.4: 17th of September 2019

- Still may corrupt the play-field. Narrowed it down to the Go-Routines 
removing completed lines if it coincides with the tick to refresh display.
	- Fixed by having lines deleted synced to next tick.

#### v.1.0.3: 17th of September 2019

- Occasionally do not detect new brick reached top and game should end...
	- Turns out it did, it just did not update the output. Fixed.
- Added score to the end-screen if new entry to top five.

#### v.1.0.2: 16th of September 2019

- Some strange formatting of the play-field at around the 2000 points mark...
	- Fixed by adjusting timings a tad.
- Got to close some input (keyboard) channels when not in relevant view they 
should be active in...
	- Fixed by giving each Event listener a unique name.
- Made the game a bit harder by lowering ticker interval and adjusting minimum 
ticks to move down.
- New Brick should now spawn in middle.

#### v.1.0.1: 16th of September 2019

- Removed some debug output

#### v.1.0.0: 16th of September 2019

- Initial release 

---

## Known issues

- N/A

## License: MIT

**Copyright © 2019 Roy Dybing** 

Permission is hereby granted, free of charge, to any person obtaining a copy of 
this software and associated documentation files (the "Software"), to deal in 
the Software without restriction, including without limitation the rights to 
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies 
of the Software, and to permit persons to whom the Software is furnished to do 
so, subject to the following conditions: The above copyright notice and this 
permission notice shall be included in all copies or substantial portions of 
the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR 
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, 
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE 
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER 
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, 
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE 
SOFTWARE.

---

ʕ◔ϖ◔ʔ