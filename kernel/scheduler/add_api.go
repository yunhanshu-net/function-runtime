package scheduler

//func (s *Coder) AddApi(api *coder.AddApiReq) error {
//	newRunner := runner.NewRunner(*api.Runner)
//	addApi, err := newRunner.AddApi(api.CodeApi)
//	//todo 这里要调用生命周期函数
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s *Coder) AddApis(api *coder.AddApisReq) (errs []*coder.CodeApiCreateInfo, err error) {
//	newRunner := runner.NewRunner(*api.Runner)
//	errs, err = newRunner.AddApis(api.CodeApis)
//	//todo 这里要调用生命周期函数
//	if err != nil {
//		return nil, err
//	}
//	return errs, nil
//}
//
//func (s *Coder) AddBizPackage(api *coder.BizPackage) (err error) {
//	newRunner := runner.NewRunner(*api.Runner)
//	err = newRunner.AddBizPackage(api)
//	//todo 这里要调用生命周期函数
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s *Coder) CreateProject(r *model.Runner) error {
//	newRunner := runner.NewRunner(*r)
//	err := newRunner.CreateProject()
//	//todo 这里要调用生命周期函数
//	if err != nil {
//		return err
//	}
//	return nil
//}
