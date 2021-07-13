package adapter

type {{ .longName }}AdapterSDK struct {
	// TODO: implement your own implementation of adapter
}

func New{{ .longName }}Adapter() ({{ .longName }}Adapter, error) {
    return {{ .longName }}AdapterSDK{}, nil
}