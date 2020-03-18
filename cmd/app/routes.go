package app


func (s *Server) InitRoutes() {

	s.router.GET(
		"/",
		s.handleIndex(),
	)
	s.router.POST(
		"/save",
		s.handleSaveFiles(),
	)
	s.router.GET(
		"/file/{id}",
		s.handleGetFile(),
	)

}
