import {RECEIVED_TICKETS} from '../actions/tickets';

const initialState = {
    list: [],
    loading: false,
    error: null
};

export default function tickets(state = initialState, action) {
    switch (action.type) {
    case RECEIVED_TICKETS:
        return {...state, list: action.data, loading: false};
    case 'TICKETS_REQUEST':
        return {...state, loading: true};
    case 'TICKETS_FAILURE':
        return {...state, loading: false, error: action.error};
    default:
        return state;
    }
}
