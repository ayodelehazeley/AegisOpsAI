from fastapi import APIRouter,Depends

from sqlalchemy.orm import Session

from app.database.database import get_db

from app.models.incident import Incident

from app.schemas.incident import IncidentCreate



router=APIRouter(
    prefix="/incidents",
    tags=["Incidents"]
)



@router.post("/")
def create_incident(
    incident:IncidentCreate,
    db:Session=Depends(get_db)
):


    new_incident=Incident(

        title=incident.title,

        description=incident.description,

        category="Unclassified",

        priority="Pending AI Analysis"

    )


    db.add(new_incident)

    db.commit()

    db.refresh(new_incident)


    return new_incident



@router.get("/")
def get_incidents(
    db:Session=Depends(get_db)
):

    return db.query(
        Incident
    ).all()