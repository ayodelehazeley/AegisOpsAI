from app.database.database import engine,Base

from app.models.incident import Incident
from app.models.asset import Asset



def create_tables():

    Base.metadata.create_all(
        bind=engine
    )


if __name__=="__main__":

    create_tables()