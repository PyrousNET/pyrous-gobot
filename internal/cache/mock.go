package cache

type MockCache struct {
}

func (m MockCache) Put(key string, value interface{}) {

}
func (m MockCache) PutAll(map[string]interface{}) {

}
func (m MockCache) Get(key string) (interface{}, bool, error) {
	return nil, false, nil
}
func (m MockCache) GetAll(keys []string) map[string]interface{} {
	return nil
}
func (m MockCache) Clean(key string) {

}
func (m MockCache) GetKeys(prefix string) ([]string, error) {
	return []string{""}, nil
}
func (m MockCache) CleanAll() {

}
