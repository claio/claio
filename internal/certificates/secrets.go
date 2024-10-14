package certificates

import "github.com/k0kubun/pp"

func CreateSecrets() error {
	certs, err := createCerts()
	if err != nil {
		return err
	}
	data := make(map[string]string)
	for _, cert := range certs {
		data[cert.Name+".crt"] = cert.CertPEM
		data[cert.Name+".key"] = cert.KeyPEM
	}
	pp.Print(data)
	return nil
}

func createCerts() ([]Certificate, error) {
	names := []string{"ca", "apiserver"} // "ca" must be the first element
	certs := make([]Certificate, 0)
	for _, name := range names {
		var cert *Certificate
		var err error
		if len(certs) > 0 {
			cert, err = CreateCertificate(name, &certs[0])
		} else {
			cert, err = CreateCertificate(name, nil)
		}
		if err != nil {
			return nil, err
		}
		certs = append(certs, *cert)
	}
	return certs, nil
}
