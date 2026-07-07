from sqlalchemy import Column,Integer,String
from app.database.database import Base



class Asset(Base):

    __tablename__="assets"


    id=Column(
        Integer,
        primary_key=True
    )


    name=Column(
        String(100)
    )


    asset_type=Column(
        String(100)
    )


    status=Column(
        String(50),
        default="Healthy"
    )


    ip_address=Column(
        String(50)
    )