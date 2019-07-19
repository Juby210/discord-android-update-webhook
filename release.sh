echo "Builiding DAUW"
go build -o build/DAUW
echo "Builiding DAUW.exe"
GOOS=windows go build -o build/DAUW.exe
echo "Builiding DAUW_freebsd"
GOOS=freebsd go build -o build/DAUW_freebsd

mkdir -p release
echo "Creating DAUW_linux_amd64.zip"
cp build/DAUW DAUW
zip -r release/DAUW_linux_amd64.zip DAUW config.example.json
rm DAUW
echo "Creating DAUW_windows_amd64.zip"
cp build/DAUW.exe DAUW.exe
zip -r release/DAUW_windows_amd64.zip DAUW.exe config.example.json
rm DAUW.exe
echo "Creating DAUW_freebsd_amd64.zip"
cp build/DAUW_freebsd DAUW_freebsd
zip -r release/DAUW_freebsd_amd64.zip DAUW_freebsd config.example.json
rm DAUW_freebsd
