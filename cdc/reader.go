package cdc

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"net"
)

type Reader struct {
	Address  string
	User     string
	Password string
	Database string
	Table    string
	Format   string // JSON or AVRO
}

// Read all cdc and send them to the given channel
// https://mariadb.com/resources/blog/how-to-stream-change-data-through-mariadb-maxscale-using-cdc-api/
func (c *Reader) Read(ctx context.Context, ch chan<- map[string]interface{}) error {
	defer close(ch)

	glog.V(2).Infof("connect to cdc %s", c.Address)
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", c.Address)
	if err != nil {
		return errors.Wrapf(err, "connect to %s failed", c.Address)
	}
	defer conn.Close()

	err = c.writeAuth(conn)
	if err != nil {
		return err
	}
	err = c.expectResponse(conn, []byte("OK"))
	if err != nil {
		return err
	}
	glog.V(1).Infof("login successful to %s", c.Address)

	id := uuid.New().String()
	_, err = fmt.Fprintf(conn, "REGISTER UUID=%s, TYPE=%s", id, c.Format)
	if err != nil {
		return err
	}
	err = c.expectResponse(conn, []byte("OK"))
	if err != nil {
		return err
	}
	glog.V(1).Infof("registered with uuid %s successful", id)

	// REQUEST-DATA DATABASE.TABLE[.VERSION] [GTID]
	_, err = fmt.Fprintf(conn, "REQUEST-DATA %s.%s", c.Database, c.Table)
	if err != nil {
		return err
	}

	errs := make(chan error)
	glog.V(1).Infof("start streaming of %s %s", c.Database, c.Table)
	go func() {
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadBytes('\n')
			if err == io.EOF {
				glog.V(1).Infof("connection closed")
				errs <- nil
			}
			if err != nil {
				errs <- errors.Wrap(err, "read line failed")
			}
			if glog.V(3) {
				glog.Infof("read %s", string(line))
			}
			var data map[string]interface{}
			if err = json.NewDecoder(bytes.NewBuffer(line)).Decode(&data); err != nil {
				errs <- errors.Wrap(err, "decode json failed")
			}
			ch <- data
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errs:
		return err
	}
}

func (c *Reader) expectResponse(conn io.Reader, response []byte) error {
	buf, err := c.read(conn)
	if err != nil {
		return err
	}
	if !bytes.Contains(buf, response) {
		return errors.New("login failed")
	}
	return nil
}

func (c *Reader) read(conn io.Reader) ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "read failed")
	}
	return buf[0:n], nil
}

func (c *Reader) writeAuth(conn io.Writer) error {
	h := sha1.New()
	io.WriteString(h, c.Password)

	encoder := hex.NewEncoder(conn)
	_, err := encoder.Write([]byte(fmt.Sprintf("%s:", c.User)))
	if err != nil {
		return errors.Wrap(err, "hex encode failed")
	}
	_, err = encoder.Write(h.Sum(nil))
	if err != nil {
		return errors.Wrap(err, "write failed")
	}
	return nil
}
