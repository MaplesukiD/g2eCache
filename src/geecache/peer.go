package geecache

import pb "goCache/src/geecache/geecachepb"

//节点选择

type PeerPicker interface {
	PeerPicker(key string) (peer PeerGetter, ok bool)
}

//	type PeerGetter interface {
//		Get(group string, key string) ([]byte, error)
//	}
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
