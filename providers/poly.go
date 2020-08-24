/*
 * Copyright (C) 2019 Nethesis S.r.l.
 * http://www.nethesis.it - info@nethesis.it
 *
 * This file is part of Falconieri project.
 *
 * Falconieri is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License,
 * or any later version.
 *
 * Falconieri is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Falconieri.  If not, see LICENSE.
 *
 * author: Matteo Valentini <matteo.valentini@nethesis.it>
 */
/* Notes from adding Polycom (might be better located in the ReadMe)
 * Before using ZTP you must add a "Profile" in the ZTP management users account
 * It is assumed you are adding device types that fit the "Polycom_UCS_Device" type
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
	"github.com/nethesis/falconieri/models"
	"github.com/nethesis/falconieri/utils"
)

type PolyDevice struct {
	Mac        string
	ServerName string
	Url        string
	Override   string
}

func (s PolyDevice) Register() error {

	var response_regexp = `(?s)<params><param><value><array><data><value><boolean>(.*)</boolean></value><value>(.*)</value></data></array></value></param></params>`

	buf, _ := xml.EncodeClientRequest("redirect.registerDeviceWithUniqueUrl", &s)

	req, _ := http.NewRequest("POST", configuration.Config.Providers.Poly.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Poly.User,
		configuration.Config.Providers.Poly.Password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return models.ProviderError{
			Message:      "connection_to_remote_provider_failed",
			WrappedError: err,
		}

	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("provider_remote_call_failed")
	}

	respBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return errors.New("read_remote_response_failed")
	}

	re := regexp.MustCompile(response_regexp)
	if re.MatchString(string(respBytes)) {

		response := re.FindStringSubmatch(string(respBytes))

		var success bool

		success, _ = strconv.ParseBool(response[1])
		message := response[2]

		if !success {
			return utils.ParseProviderError(message)
		}

	} else {
		return errors.New("unknown_response_from_provider")
	}

	return nil

}