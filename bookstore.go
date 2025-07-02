package main

import (
	"bookstore_enhance/proto"
	"bookstore_enhance/proto/protoconnect"
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

const (
	defaultCursor   = "0" // 默认游标
	defaultPageSize = 2   // 默认每页显示数量
)

type server struct {
	protoconnect.UnimplementedBookStoreHandler
	bs *bookstore
}

func (s *server) ListShelves(ctx context.Context, in *connect.Request[emptypb.Empty]) (*connect.Response[proto.ListShelvesResponse], error) {
	sl, err := s.bs.ListShelves(ctx)
	if err == gorm.ErrEmptySlice {
		return connect.NewResponse(&proto.ListShelvesResponse{}), nil
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	nsl := make([]*proto.Shelf, 0, len(sl))
	for _, s := range sl {
		nsl = append(nsl, &proto.Shelf{
			Id:    s.ID,
			Theme: s.Theme,
			Size:  s.Size,
		})
	}
	return connect.NewResponse(&proto.ListShelvesResponse{
		Shelves: nsl,
	}), nil
}

func (s *server) CreateShelf(ctx context.Context, in *connect.Request[proto.CreateShelfRequest]) (*connect.Response[proto.Shelf], error) {
	if len(in.Msg.GetShelf().GetTheme()) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("theme is required"))

	}
	data := Shelf{
		Theme: in.Msg.GetShelf().GetTheme(),
		Size:  in.Msg.GetShelf().GetSize(),
	}
	ns, err := s.bs.CreateShelf(ctx, data)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&proto.Shelf{
		Id:    ns.ID,
		Theme: ns.Theme,
		Size:  ns.Size,
	}), nil
}

func (s *server) GetShelf(ctx context.Context, in *connect.Request[proto.GetShelfRequest]) (*connect.Response[proto.Shelf], error) {
	if in.Msg.GetShelf() <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("shelf ID is required"))
	}
	shelf, err := s.bs.GetShelf(ctx, in.Msg.GetShelf())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&proto.Shelf{
		Id:    shelf.ID,
		Theme: shelf.Theme,
		Size:  shelf.Size,
	}), nil
}

func (s *server) DeleteShelf(ctx context.Context, in *connect.Request[proto.DeleteShelfRequest]) (*connect.Response[emptypb.Empty], error) {
	if in.Msg.GetShelf() <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("shelf ID is required"))
	}
	err := s.bs.DeleteShelf(ctx, in.Msg.GetShelf())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *server) ListBooks(ctx context.Context, in *connect.Request[proto.ListBooksRequest]) (*connect.Response[proto.ListBooksResponse], error) {
	// 参数查询
	if in.Msg.GetShelf() <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid shelf ID"))
	}
	// 没有分页 token 默认第一页
	var (
		cursor       = defaultCursor
		pageSize int = defaultPageSize
	)
	log.Printf("Page token: %s", in.Msg.GetPageToken())
	if len(in.Msg.GetPageToken()) > 0 {
		pageInfo := Token(in.Msg.GetPageToken()).Decode()
		// 有分页先解析分页数据
		if pageInfo.InValid() {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page token"))
		}
		cursor = pageInfo.NextID
		pageSize = int(pageInfo.PageSize)
	}
	log.Printf("cursor: %s, pageSize: %d, shelfID: %d", cursor, pageSize, in.Msg.GetShelf())
	// 查询数据库, 基于游标实现分页
	bookList, err := s.bs.GetBookListByShelfID(ctx, in.Msg.GetShelf(), cursor, pageSize+1)
	if err != nil {
		log.Printf("failed to get book list: %v", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get book list"))
	}
	// 如果查询出来的结果比 pageSize 大，，那么说明有下一页
	var (
		hasNextPage   bool
		nextPageToken string
		realSize      int = len(bookList)
	)
	// 当查询数据库的结果数大于 pageSize
	if len(bookList) > pageSize {
		// 说明有下一页
		hasNextPage = true
		// 下面格式化数据没必要把所有查询结果都返回，只需要返回 pageSize 个数据
		realSize = pageSize
	}
	// 封装返回的数据
	res := make([]*proto.Book, 0, len(bookList))
	for i := range realSize {
		res = append(res, &proto.Book{
			Id:     bookList[i].ID,
			Author: bookList[i].Author,
			Title:  bookList[i].Title,
		})
	}

	// 如果有下一页，，生成下一页的 token
	if hasNextPage {
		nextPageInfo := &Page{
			NextID:        strconv.FormatInt(res[realSize-1].Id, 10), // res[realSize - 1].Id 最后一个返回结果的 id
			NextTimeAtUTC: time.Now().Unix(),
			PageSize:      int64(pageSize),
		}
		nextPageToken = string(nextPageInfo.Encode())
	}
	return connect.NewResponse(&proto.ListBooksResponse{
		Books:         res,
		NextPageToken: nextPageToken,
	}), nil
}

func (s *server) CreateBook(ctx context.Context, in *connect.Request[proto.CreateBookRequest]) (*connect.Response[proto.Book], error) {
	if in.Msg.GetShelf() <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("shelf ID is required"))
	}
	if len(in.Msg.GetBook().Author) == 0 || len(in.Msg.GetBook().Title) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("author or title is required"))
	}
	data := &Book{
		Author:  in.Msg.GetBook().GetAuthor(),
		Title:   in.Msg.GetBook().GetTitle(),
		ShelfID: in.Msg.GetShelf(),
	}
	ns, err := s.bs.CreateBook(ctx, data)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&proto.Book{
		Id:     ns.ID,
		Author: ns.Author,
		Title:  ns.Title,
	}), nil

}

func (s *server) GetBook(context.Context, *connect.Request[proto.GetBookRequest]) (*connect.Response[proto.Book], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bookstore.BookStore.GetBook is not implemented"))
}

func (s *server) DeleteBook(context.Context, *connect.Request[proto.DeleteBookRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("bookstore.BookStore.DeleteBook is not implemented"))
}
