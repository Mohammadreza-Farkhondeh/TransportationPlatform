from fastapi import APIRouter, Depends, HTTPException, Depends
from sqlalchemy.orm import Session
from app.api.deps import get_db
from app.api.schemas import RideCreate, RideUpdate, Ride
from app.crud.ride import ride_crud
from app.services.producer import produce_ride_request_message
from app.api.deps import get_current_user_id, get_ride_id_by_user

router = APIRouter()

from fastapi import BackgroundTasks, HTTPException


@router.post("/ride/", response_model=Ride)
async def create_ride(
    ride_in: RideCreate,
    background_tasks: BackgroundTasks,
    db: Session = Depends(get_db),
    user_id: str = Depends(get_current_user_id),
):
    ride_in.passenger_id = user_id
    ride = ride_crud.create(db, obj_in=ride_in)
    background_tasks.add_task(produce_ride_request_message, ride_id=ride.id)
    return ride


@router.put("/rides/{ride_id}/accept", response_model=Ride)
async def accept_ride(
    ride_id: int,
    db: Session = Depends(get_db),
    user_id: str = Depends(get_current_user_id),
):
    # Check if the user is a driver (You might need to implement this check)
    if not is_user_driver(user_id):  # Implement this function to check user's role
        raise HTTPException(status_code=403, detail="Only drivers can accept rides")

    ride = ride_crud.get(db, id=ride_id)
    if ride is None:
        raise HTTPException(status_code=404, detail="Ride not found")

    # Update ride status and driver ID
    ride.status = RideStatus.ACCEPTED
    ride.driver_id = user_id
    db.commit()

    return ride


@router.put("/rides/{ride_id}", response_model=Ride)
async def update_ride(
    ride_id: int,
    ride_in: RideUpdate,
    db: Session = Depends(get_db),
    user_id: str = Depends(get_current_user_id),
):
    active_ride_id = get_ride_id_by_user(db, user_id)
    if active_ride_id != ride_id:
        raise HTTPException(
            status_code=403, detail="You can only update your active ride"
        )

    ride = ride_crud.get(db, id=ride_id)
    if ride is None:
        raise HTTPException(status_code=404, detail="Ride not found")

    # Update ride status and other fields as needed
    ride.status = ride_in.status
    # Update other fields here
    db.commit()

    return ride


@router.post("/rides/{ride_id}/cancel")
async def cancel_ride(
    ride_id: int,
    user_id: str = Depends(get_current_user_id),
    db: Session = Depends(get_db),
):
    active_ride_id = get_ride_id_by_user(db, user_id)
    if active_ride_id != ride_id:
        raise HTTPException(
            status_code=403, detail="You can only cancel your active ride"
        )

    ride = ride_crud.get(db, id=ride_id)
    if ride is None:
        raise HTTPException(status_code=404, detail="Ride not found")

    # Implement your cancellation logic here
    ride.status = RideStatus.CANCELED
    db.commit()

    return {"message": f"Ride {ride_id} canceled by user {user_id}"}
