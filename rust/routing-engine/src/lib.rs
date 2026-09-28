use std::os::raw::c_int;

/// Calculates a routing score based on latency, load, and error rate.
/// A lower score is better.
#[no_mangle]
pub extern "C" fn calculate_route_score(latency_ms: c_int, load_factor: c_int, error_rate: c_int) -> c_int {
    // Basic weighted scoring algorithm
    let weight_latency = 2;
    let weight_load = 5;
    let weight_error = 10;

    (latency_ms * weight_latency) + (load_factor * weight_load) + (error_rate * weight_error)
}
