package jwt


func (t *token) Uid() (result string, err error) {
	return t.uid, nil
}

func (t *token) SetUid(value string) (err error) {
	t.uid = value
	return
}

func (t *token) Name() (result string, err error) {
	return t.info.name, nil
}

func (t *token) SetName(value string) (err error) {
	t.info.name = value
	return
}