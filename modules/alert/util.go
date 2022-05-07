package alert

import (
	"fmt"

	"github.com/Leantar/fimserver/models"
)

func GetDifference(obj1, obj2 models.FsObject) (difference string) {
	if obj1.Hash != obj2.Hash {
		difference = fmt.Sprintf("hash: %s -> %s", obj1.Hash, obj2.Hash)
	} else if obj1.Uid != obj2.Uid || obj1.Gid != obj2.Gid {
		difference = fmt.Sprintf("owner: %d:%d -> %d:%d", obj1.Uid, obj1.Gid, obj2.Uid, obj2.Gid)
	} else if obj1.Mode != obj2.Mode {
		difference = fmt.Sprintf("mode: %s -> %s", obj1.ParseMode(), obj2.ParseMode())
	} else if obj1.Created != obj2.Created {
		difference = fmt.Sprintf("creation time: %d -> %d", obj1.Created, obj2.Created)
	} else if obj1.Modified != obj2.Modified {
		difference = fmt.Sprintf("modified: %d -> %d", obj1.Modified, obj2.Modified)
	}

	return
}
