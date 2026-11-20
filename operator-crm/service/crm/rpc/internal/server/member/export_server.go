package server

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	logic "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/logic/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	pb "oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/memberexportpb"
)

type ExportServer struct {
	pb.UnimplementedMemberExportServer
	Svc *svc.ServiceContext
}

func (s *ExportServer) Export(in *pb.ExportRequest, stream grpc.ServerStreamingServer[pb.ExportChunk]) error {
	return logic.StreamMembers(stream.Context(), s.Svc, in, func(ctx context.Context, chunk *pb.ExportChunk) error {
		// At most one transport send is in flight per request. Returning on deadline
		// closes the RPC transport and unblocks its send; this is not a background job.
		result := make(chan error, 1)
		go func() { result <- stream.Send(chunk) }()
		select {
		case err := <-result:
			return err
		case <-ctx.Done():
			return status.FromContextError(ctx.Err()).Err()
		}
	})
}
