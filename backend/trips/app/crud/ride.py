from typing import List

from sqlalchemy.orm import Session

from app.db.models import Ride
from app.crud.base import CRUDBase
from app.api.schemas import RideCreate, RideUpdate


class RideCRUD(CRUDBase[Ride, RideCreate, RideUpdate]):
    def get_by_passenger_id(
        self, db: Session, passenger_id: str, skip: int = 0, limit: int = 100
    ) -> List[Ride]:
        return (
            db.query(self.model)
            .filter(self.model.passenger_id == passenger_id)
            .offset(skip)
            .limit(limit)
            .all()
        )

    def get_by_driver_id(
        self, db: Session, driver_id: str, skip: int = 0, limit: int = 100
    ) -> List[Ride]:
        return (
            db.query(self.model)
            .filter(self.model.driver_id == driver_id)
            .offset(skip)
            .limit(limit)
            .all()
        )


ride_crud = RideCRUD(Ride)
