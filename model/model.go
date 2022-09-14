package model

func GetTUForModel(modelName string) (string, error) {
	return "entity", nil
}

func GetAllTUs() (string, error) {
	return "all entities", nil
}

func UpdateTUForModel(modelName string, throughputUnit int) error {
	return nil
}
