// Copyright 2017 Matt Martz <matt@sivel.net>
// All Rights Reserved.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

const html string = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Perky</title>
    <link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
    <!--<link href="http://getbootstrap.com/examples/jumbotron-narrow/jumbotron-narrow.css" rel="stylesheet">-->
    <style>
      @media (min-width: 768px) {
        .container {
          max-width: 730px;
        }
      }
      .center {
        text-align: center;
      }
      .full {
        width: 100%;
      }
    </style>
    <!--[if lt IE 9]>
      <script src="https://oss.maxcdn.com/html5shiv/3.7.3/html5shiv.min.js"></script>
      <script src="https://oss.maxcdn.com/respond/1.4.2/respond.min.js"></script>
    <![endif]-->
  </head>
  <body>
    <div class="container">
      <div class="header clearfix">
        <h3 class="text-muted">Perky</h3>
      </div>
      <div class="jumbotron center">
        <!--<h1>Jumbotron heading</h1>-->
        <p>
          <form class="form-inline" action="" method="post" enctype="multipart/form-data">
            <input class="form-control input-lg" type="file" name="file">
            <button class="btn btn-lg" type="submit">Upload</button>
          </form>
        </p>
      </div>
      <div class="jumbotron">
        <h4>Files</h4>
        <table class="table-striped table-hover full">

        {{range .}}
          <tr><td><a href="{{.}}">{{.}}</a></td></tr>
        {{end}}
        </table>
      </div>
    </div>
  </body>
</html>
`

func index(c echo.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	files, err := ioutil.ReadDir(cwd)
	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, file.Name())
		}
	}

	t, err := template.New("t").Parse(html)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	t.Execute(c.Response().Writer(), fileList)

	return nil
}

func save(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	cwd, err := os.Getwd()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	dstPath := filepath.Join(cwd, file.Filename)

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func main() {
	var port string
	flag.StringVar(&port, "port", ":8000", "HOST:PORT to listen on, HOST not required to listen on all addresses")
	flag.Parse()

	if strings.HasPrefix(port, ":") {
		fmt.Printf("http://0.0.0.0%s\n", port)
	} else {
		fmt.Printf("http://%s\n", port)
	}

	//127.0.0.1 - - [06/Oct/2016 16:03:58] "GET / HTTP/1.1" 200 -

	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${remote_ip} - - [${time_rfc3339}] \"${method} ${uri} HTTP/1.1\" ${status} -\n",
	}))
	e.GET("/", index)
	e.POST("/", save)
	e.Static("/", "")
	e.Run(standard.New(port))
}
