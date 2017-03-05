#!/bin/bash

rm out.dca
youtube-dl -o out.wav -x --audio-format wav $1 
./bin/dca -i out.wav --raw -vol 4  > out.dca
rm out.wav
