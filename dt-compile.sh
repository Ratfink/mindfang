#!/bin/sh

dtc_path="$HOME"

${dtc_path}/dtc -O dtb -o BB-MINDFANG-00A0.dtbo -b 0 -@ BB-MINDFANG-00A0.dts
