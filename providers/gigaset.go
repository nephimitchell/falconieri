/*
 * Copyright (C) 2019 Nethesis S.r.l.
 * http://www.nethesis.it - info@nethesis.it
 *
 * This file is part of Falconieri project.
 *
 * Icaro is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License,
 * or any later version.
 *
 * Icaro is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Icaro.  If not, see LICENSE.
 *
 * author: Matteo Valentini <matteo.valentini@nethesis.it>
 */

package providers

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/divan/gorilla-xmlrpc/xml"

	"github.com/nethesis/falconieri/configuration"
)

type GigasetDevice struct {
	Mac      string
	Url      string
	Provider string
}

func (d GigasetDevice) Register() error {

	var response_regexp = `<params><param><value><array><data><value><boolean>(.*)</boolean></value><value>(.*)</value></data></array></value></param></params>`

	//Deregister the device, dont care if the device already exists or not
	buf, _ := xml.EncodeClientRequest("autoprov.deregisterDevice", &struct{ Mac string }{Mac: d.Mac})

	req, _ := http.NewRequest("POST", configuration.Config.Providers.Gigaset.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Gigaset.User,
		configuration.Config.Providers.Gigaset.Password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Remote call failed")
	}

	//Register the device
	buf, _ = xml.EncodeClientRequest("autoprov.registerDevice", &d)

	req, _ = http.NewRequest("POST", configuration.Config.Providers.Gigaset.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Gigaset.User,
		configuration.Config.Providers.Gigaset.Password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err = http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Remote call failed")
	}

	respBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return errors.New("Error on reading Gigaset response")
	}

	re := regexp.MustCompile(response_regexp)

	if re.MatchString(string(respBytes)) {

		response := re.FindStringSubmatch(string(respBytes))

		var success bool

		success, _ = strconv.ParseBool(response[1])
		message := response[2]

		if !success {
			return errors.New("Error to register Device on Gigaset provider: " + message)
		}

	} else {
		return errors.New("Unknow response from Gigaset provider")
	}

	return nil
}