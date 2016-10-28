package routers

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"time"
)

const (
	ETCD_CLIENT_TIMEOUT = 3 * time.Second

	ROLE_LEADER   = "leader"
	ROLE_FOLLOWER = "follower"

	STATUS_HEALTHY   = "healthy"
	STATUS_UNHEALTHY = "unhealthy"
)

type Member struct {
	*etcdserverpb.Member
	Role   string `json:"role"`
	Status string `json:"status"`
	DbSize int64  `json:"db_size"`
}

func newEtcdCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), ETCD_CLIENT_TIMEOUT)
	return ctx
}

func getMembersHandler(client *clientv3.Client) respHandler {
	return func(c *gin.Context) (interface{}, error) {
		resp, err := client.MemberList(newEtcdCtx())
		if err != nil {
			return nil, err
		}

		members := []*Member{}
		for _, member := range resp.Members {
			if len(member.ClientURLs) > 0 {
				m := &Member{Member: member, Role: ROLE_FOLLOWER, Status: STATUS_UNHEALTHY}
				resp, err := client.Status(newEtcdCtx(), m.ClientURLs[0])
				if err == nil {
					m.Status = STATUS_HEALTHY
					m.DbSize = resp.DbSize
					if resp.Leader == resp.Header.MemberId {
						m.Role = ROLE_LEADER
					}
				}
				members = append(members, m)
			}
		}

		return members, nil
	}
}
