package model

import (
	"fiy/app/cmdb/models/model"
	orm "fiy/common/global"
	"fiy/tools/app"

	"github.com/gin-gonic/gin"
)

/*
  @Author : lanyulei
*/

// 创建模型
func CreateModelInfo(c *gin.Context) {
	var (
		err       error
		info      model.Info
		infoCount int64
	)

	err = c.ShouldBind(&info)
	if err != nil {
		app.Error(c, -1, err, "参数绑定失败")
		return
	}

	// 查询分组是否存在， 分组唯一标识及名称都不存在，才可创建分组
	info.IsUsable = true
	err = orm.Eloquent.
		Model(&info).
		Where("identifies = ? or name = ?", info.Identifies, info.Name).
		Count(&infoCount).Error
	if err != nil {
		app.Error(c, -1, err, "查询模型是否存在失败")
		return
	}
	if infoCount > 0 {
		app.Error(c, -1, nil, "模型唯一标识或名称已存在")
		return
	}

	// 写入数据库
	err = orm.Eloquent.Create(&info).Error
	if err != nil {
		app.Error(c, -1, err, "创建模型失败")
		return
	}

	app.OK(c, nil, "")
}

// 创建模型字段分组
func CreateModelFieldGroup(c *gin.Context) {
	var (
		err             error
		fieldGroup      model.FieldGroup
		fieldGroupCount int64
	)

	err = c.ShouldBind(&fieldGroup)
	if err != nil {
		app.Error(c, -1, err, "参数绑定失败")
		return
	}

	// 验证字段分组是否存在
	err = orm.Eloquent.Model(&fieldGroup).Where("name = ?", fieldGroup.Name).Count(&fieldGroupCount).Error
	if err != nil {
		app.Error(c, -1, err, "查询字段分组是否存在失败")
		return
	}
	if fieldGroupCount > 0 {
		app.Error(c, -1, nil, "字段分组名称已存在，请确认")
		return
	}

	// 创建字段分组
	err = orm.Eloquent.Create(&fieldGroup).Error
	if err != nil {
		app.Error(c, -1, err, "创建字段分组失败")
		return
	}

	app.OK(c, nil, "")
}

// 获取模型详情
func GetModelDetails(c *gin.Context) {
	var (
		err          error
		fieldDetails struct {
			model.Info
			FieldGroups []*struct {
				model.FieldGroup
				Fields []*model.Fields `json:"fields"`
			} `json:"field_groups"`
		}
		modelId string
	)

	modelId = c.Param("id")

	// 查询模型信息
	err = orm.Eloquent.Model(&model.Info{}).Where("id = ?", modelId).Find(&fieldDetails).Error
	if err != nil {
		app.Error(c, -1, err, "查询模型信息失败")
		return
	}

	// 查询模型分组
	err = orm.Eloquent.Model(&model.FieldGroup{}).Where("info_id = ?", modelId).Find(&fieldDetails.FieldGroups).Error
	if err != nil {
		app.Error(c, -1, err, "查询模型信息失败")
		return
	}

	// 获取分组对应的字段
	for _, group := range fieldDetails.FieldGroups {
		err = orm.Eloquent.Model(&model.Fields{}).
			Where("info_id = ? and field_group_id = ?", modelId, group.Id).
			Find(&group.Fields).Error
		if err != nil {
			app.Error(c, -1, err, "查询字段列表失败")
			return
		}
	}

	app.OK(c, fieldDetails, "")
}

// 创建模型字段
func CreateModelField(c *gin.Context) {
	var (
		err        error
		fieldValue model.Fields
		fieldCount int64
	)

	err = c.ShouldBind(&fieldValue)
	if err != nil {
		app.Error(c, -1, err, "参数绑定失败")
		return
	}

	// 判断唯一标识及名称是否唯一
	err = orm.Eloquent.
		Model(&model.Fields{}).
		Where("info_id = ? and (identifies = ? or name = ?)", fieldValue.InfoId, fieldValue.Identifies, fieldValue.Name).
		Count(&fieldCount).Error
	if err != nil {
		app.Error(c, -1, err, "验证唯一标识或者名称的唯一性失败")
		return
	}
	if fieldCount > 0 {
		app.Error(c, -1, nil, "唯一标识或者名称出现重复，请确认。")
		return
	}

	// 创建字段
	err = orm.Eloquent.Create(&fieldValue).Error
	if err != nil {
		app.Error(c, -1, err, "创建字段失败")
		return
	}

	app.OK(c, nil, "")
}

// 更新模型字段
func EditModelField(c *gin.Context) {
	var (
		err     error
		field   model.Fields
		fieldId string
	)

	fieldId = c.Param("id")

	err = c.ShouldBind(&field)
	if err != nil {
		app.Error(c, -1, err, "参数绑定失败")
		return
	}

	err = orm.Eloquent.Model(&field).Where("id = ?", fieldId).Updates(map[string]interface{}{
		"identifies":    field.Identifies,
		"name":          field.Name,
		"type":          field.Type,
		"is_edit":       field.IsEdit,
		"is_unique":     field.IsUnique,
		"prompt":        field.Prompt,
		"configuration": field.Configuration,
	}).Error
	if err != nil {
		app.Error(c, -1, err, "参数绑定失败")
		return
	}

	app.OK(c, nil, "")
}

// 删除模型分组
func DeleteFieldGroup(c *gin.Context) {
	var (
		err          error
		fieldGroupId string
		fieldCount   int64
	)

	fieldGroupId = c.Param("id")

	// 如果分组下有对应字段，则无法删除
	err = orm.Eloquent.Model(&model.Fields{}).Where("field_group_id = ?", fieldGroupId).Count(&fieldCount).Error
	if err != nil {
		app.Error(c, -1, err, "查询字段列表失败")
		return
	}
	if fieldCount > 0 {
		app.Error(c, -1, err, "无法删除分组，因分组下有对应的字段数据")
		return
	}

	// 删除字段分组
	err = orm.Eloquent.Delete(&model.FieldGroup{}, fieldGroupId).Error
	if err != nil {
		app.Error(c, -1, err, "删除字段分组失败")
		return
	}

	app.OK(c, nil, "")
}

// 编辑字段分组
func EditFieldGroup(c *gin.Context) {
	var (
		err          error
		fieldGroup   model.FieldGroup
		fieldGroupId string
	)

	fieldGroupId = c.Param("id")

	err = c.ShouldBind(&fieldGroup)
	if err != nil {
		app.Error(c, -1, err, "参数绑定失败")
		return
	}

	err = orm.Eloquent.Model(&fieldGroup).Where("id = ?", fieldGroupId).Updates(map[string]interface{}{
		"name":     fieldGroup.Name,
		"sequence": fieldGroup.Sequence,
		"is_fold":  fieldGroup.IsFold,
	}).Error
	if err != nil {
		app.Error(c, -1, err, "更新字段分组失败")
		return
	}

	app.OK(c, nil, "")
}
