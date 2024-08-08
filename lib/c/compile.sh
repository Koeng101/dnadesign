go build -o libawesome.so -buildmode=c-shared lib.go
gcc testlib.c -o testlib -L. -lawesome -Wl,-rpath='$ORIGIN'
./testlib
rm testlib
python3 test.py
