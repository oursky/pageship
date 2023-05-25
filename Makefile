.PHONY: migration
migration:
ifndef NAME
	$(error NAME is undefined)
endif
	migrate create -ext sql -dir migrations/sqlite -seq $(NAME)
	migrate create -ext sql -dir migrations/postgres -seq $(NAME)
