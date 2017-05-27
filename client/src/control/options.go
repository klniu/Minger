package control

import . "Minger/client/src/model"

func (d *DB) ReadOption(key string) (Option, error) {
    var option Option
    if err := d.Where("name=?", key).First(&option).Error; err != nil {
        return option, err
    }
    return option, nil
}
