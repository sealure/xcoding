package handler

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 分页每页最大值上限，避免过大查询占用资源
const maxPageSize = 100

// normalizePagination 统一处理分页参数的默认值与边界校验
func normalizePagination(page, pageSize int32) (int32, int32, error) {
	if page == 0 {
		page = 1
	}
	if page < 1 {
		return 0, 0, status.Errorf(codes.InvalidArgument, "page must be >= 1")
	}

	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize < 1 {
		return 0, 0, status.Errorf(codes.InvalidArgument, "page_size must be >= 1")
	}
	if pageSize > maxPageSize {
		return 0, 0, status.Errorf(codes.InvalidArgument, "page_size must be <= %d", maxPageSize)
	}

	return page, pageSize, nil
}
