---
rm -r ../thispage_backup;
cp -r . ../thispage_backup;
pull --prepend pkg;
pull --prepend main.go;
inp move 1789 169;
inp click;
inp paste;
inp wait 1000;
inp press enter;
inp mov 400 400;
---








