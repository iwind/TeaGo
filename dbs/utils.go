// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbs

func anyError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
