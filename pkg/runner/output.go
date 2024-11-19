package runner

import (
	"fmt"
	"github.com/wjlin0/acmeGoBaidu/pkg/certificate"
)

func (r *Runner) Output() error {
	// 保存更新后的证书信息
	err := certificate.SaveCertificateInfo(r.Certificates, r.JsonFilePath)
	if err != nil {
		return fmt.Errorf("保存证书信息失败: %v", err)
	}
	return nil
}
