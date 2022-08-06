package core

func (db *Database) Setup() {
	db.accepting = true
}

func (db *Database) Start() {
	var request DatabaseRequest
	for db.accepting {
		request = <-db.C1 // request from http_server
		db.Handler(request)

		request = <-db.C7 // request from supervisor
		db.Handler(request)
	}

	db.wg.Wait()
}

func (db Database) Handler(request DatabaseRequest) {
	switch request.action {
	case Read:
		break
	case Write:
		break
	case Delete:
		break
	}
}

func (db *Database) Teardown() {
	db.accepting = false
}
