package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func updateDB() {
	serverLog("Updating database...\n")

	dbURL := "https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz"

	err := downloadFile("maxmind/GeoLite2-City.tar.gz", dbURL)
	if err != nil {
		serverLog("error updating database")
		run = false
		return
	}

	serverLog("Extracting database...\n")
	file, err := os.Open("maxmind/GeoLite2-City.tar.gz")

	gzr, err := gzip.NewReader(file)
	if err != nil {
		serverLog("error updating database")
		run = false
		return
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	waiting := true
	for waiting {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
		case err != nil:
			serverLog("error updating database")
			run = false
			return
		case header == nil:
			continue
		}

		fileName := filepath.Base(header.Name)

		if header.Typeflag == tar.TypeReg && fileName == "GeoLite2-City.mmdb" {
			serverLog("Found database in archive file\n")

			f, err := os.OpenFile(filepath.Join("maxmind/", fileName), os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				serverLog("error updating database")
				run = false
				return
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				serverLog("error updating database")
				run = false
				return
			}

			f.Close()

			waiting = false
		}
	}

	os.Remove("maxmind/GeoLite2-City.tar.gz")

	fmt.Println("Archive File Deleted")
}

func downloadFile(filepath string, url string) error {

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
