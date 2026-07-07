from sqlalchemy import Column, Integer, String, Text, DateTime
from datetime import datetime

from app.database.database import Base


class Incident(Base):

    __tablename__ = "incidents"


    id = Column(
        Integer,
        primary_key=True,
        index=True
    )


    title = Column(
        String(255),
        nullable=False
    )


    description = Column(
        Text
    )


    category = Column(
        String(100)
    )


    priority = Column(
        String(50),
        default="Medium"
    )


    status = Column(
        String(50),
        default="Open"
    )


    ai_confidence = Column(
        String(20)
    )


    created_at = Column(
        DateTime,
        default=datetime.utcnow
    )