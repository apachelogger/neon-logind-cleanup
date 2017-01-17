/*
	Copyright 2016 Harald Sitter <sitter@kde.org>

	This program is free software; you can redistribute it and/or
	modify it under the terms of the GNU General Public License as
	published by the Free Software Foundation; either version 3 of
	the License or any later version accepted by the membership of
	KDE e.V. (or its successor approved by the membership of KDE
	e.V.), which shall act as a proxy defined in Section 14 of
	version 3 of the license.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"net"
	"time"

	"github.com/godbus/dbus"

	_ "net/http/pprof"
)

type sessionEntry struct {
	ID         string
	ObjectPath dbus.ObjectPath
}

type Blacklist struct {
	Hostname string
	IPs      []net.IP
}

func NewBlacklist(hostname string) *Blacklist {
	blacklist := &Blacklist{
		Hostname: hostname,
	}
	ips, err := net.LookupIP(hostname)
	if err != nil {
		fmt.Println("failed to resolve hostname!")
		panic(err)
	}
	blacklist.IPs = ips
	return blacklist
}

func (blacklist *Blacklist) blackListed(host string) bool {
	if blacklist.Hostname == host {
		return true
	}
	for _, ip := range blacklist.IPs {
		if ip.String() == host {
			return true
		}
	}
	return false
}

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		panic(err)
	}

	userSelf := conn.Object("org.freedesktop.login1", "/org/freedesktop/login1/user/self")
	variant, err := userSelf.GetProperty("org.freedesktop.login1.User.Sessions")
	if err != nil {
		fmt.Println(err)
		panic(variant)
	}

	// Property would be a(so) which in godbus terms is a slice of of empty
	// interfaces slices. The effective length is the len of the first level.
	sessions := make([]sessionEntry, len(variant.Value().([][]interface{})))
	err = dbus.Store([]interface{}{variant.Value()}, &sessions)
	if err != nil {
		panic(err)
	}

	blacklist := NewBlacklist("drax.kde.org")

	for _, session := range sessions {
		fmt.Println("--")
		fmt.Println(session)
		sessionObject := conn.Object("org.freedesktop.login1", session.ObjectPath)
		remoteHostVariant, err := sessionObject.GetProperty("org.freedesktop.login1.Session.RemoteHost")
		if err != nil {
			fmt.Println(err)
			continue
		}
		remoteHost := remoteHostVariant.Value().(string)
		fmt.Println(remoteHost)
		if !blacklist.blackListed(remoteHost) {
			fmt.Printf("Session %s is not from offending remote.\n", session.ID)
			continue
		}

		timestampVariant, err := sessionObject.GetProperty("org.freedesktop.login1.Session.Timestamp")
		if err != nil {
			fmt.Println(err)
			continue
		}
		timestamp := timestampVariant.Value().(uint64)
		start := time.Unix(int64(timestamp/1000000), 0)
		fmt.Println(start)
		if time.Now().Sub(start).Hours() <= 0 {
			fmt.Printf("Session %s is not too old yet %s.\n", session.ID, time.Now().Sub(start))
			continue
		}

		fmt.Printf("Session %s is older than 6 hours. Terminating.\n", session.ID)
		call := sessionObject.Call("org.freedesktop.login1.Session.Terminate", 0)
		if call.Err != nil {
			fmt.Println(call.Err)
		}
	}
}
