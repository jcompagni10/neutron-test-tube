use neutron_sdk::proto_types::neutron::dex::*;

use test_tube::module::Module;
use test_tube::runner::Runner;
use test_tube::{fn_execute, fn_query};

pub struct Dex<'a, R: Runner<'a>> {
    runner: &'a R,
}

impl<'a, R: Runner<'a>> Module<'a, R> for Dex<'a, R> {
    fn new(runner: &'a R) -> Self {
        Self { runner }
    }
}

impl<'a, R> Dex<'a, R>
where
    R: Runner<'a>,
{
    fn_execute! {
        pub deposit: MsgDeposit["/neutron.dex.MsgDeposit"] => MsgDepositResponse
    }

    fn_execute! {
        pub withdrawal: MsgWithdrawal["/neutron.dex.MsgWithdrawal"] => MsgWithdrawalResponse
    }

    fn_execute! {
        pub place_limit_order: MsgPlaceLimitOrder["/neutron.dex.MsgPlaceLimitOrder"] => MsgPlaceLimitOrderResponse
    }

    fn_execute! {
        pub withdraw_filled_limit_order: MsgWithdrawFilledLimitOrder["/neutron.dex.MsgWithdrawFilledLimitOrder"] => MsgWithdrawFilledLimitOrderResponse
    }

    fn_execute! {
        pub cancel_limit_order: MsgCancelLimitOrder["/neutron.dex.MsgCancelLimitOrder"] => MsgCancelLimitOrderResponse
    }

    fn_execute! {
        pub multi_hop_swap: MsgMultiHopSwap["/neutron.dex.MsgMultiHopSwap"] => MsgMultiHopSwapResponse
    }

    fn_query! {
        pub get_params["/neutron.dex.Query/Params"]: QueryParamsRequest => QueryParamsResponse
    }

    fn_query! {
        pub get_limit_order_tranche_user["/neutron.dex.Query/LimitOrderTrancheUser"]: QueryGetLimitOrderTrancheUserRequest => QueryGetLimitOrderTrancheUserResponse
    }

    fn_query! {
        pub get_limit_order_tranche_user_all["/neutron.dex.Query/LimitOrderTrancheUserAll"]: QueryAllLimitOrderTrancheUserRequest => QueryAllLimitOrderTrancheUserResponse
    }

    fn_query! {
        pub get_limit_order_tranche_user_all_by_address["/neutron.dex.Query/LimitOrderTrancheUserAllByAddress"]: QueryAllUserLimitOrdersRequest => QueryAllUserLimitOrdersResponse
    }

    fn_query! {
        pub get_limit_order_tranche["/neutron.dex.Query/LimitOrderTranche"]: QueryGetLimitOrderTrancheRequest => QueryGetLimitOrderTrancheResponse
    }

    fn_query! {
        pub get_limit_order_tranche_all["/neutron.dex.Query/LimitOrderTrancheAll"]: QueryAllLimitOrderTrancheRequest => QueryAllLimitOrderTrancheResponse
    }

    fn_query! {
        pub get_user_deposits_all["/neutron.dex.Query/UserDepositsAll"]: QueryAllUserDepositsRequest => QueryAllUserDepositsResponse
    }

    fn_query! {
        pub get_tick_liquidity_all["/neutron.dex.Query/TickLiquidityAll"]: QueryAllTickLiquidityRequest => QueryAllTickLiquidityResponse
    }

    fn_query! {
        pub get_inactive_limit_order_tranche["/neutron.dex.Query/InactiveLimitOrderTranche"]: QueryGetInactiveLimitOrderTrancheRequest => QueryGetInactiveLimitOrderTrancheResponse
    }

    fn_query! {
        pub get_inactive_limit_order_tranche_all["/neutron.dex.Query/InactiveLimitOrderTrancheAll"]: QueryAllInactiveLimitOrderTrancheRequest => QueryAllInactiveLimitOrderTrancheResponse
    }

    fn_query! {
        pub get_pool_reserves_all["/neutron.dex.Query/PoolReservesAll"]: QueryAllPoolReservesRequest => QueryAllPoolReservesResponse
    }

    fn_query! {
        pub get_pool_reserves["/neutron.dex.Query/PoolReserves"]: QueryGetPoolReservesRequest => QueryGetPoolReservesResponse
    }

    fn_query! {
        pub get_estimate_multi_hop_swap["/neutron.dex.Query/EstimateMultiHopSwap"]: QueryEstimateMultiHopSwapRequest => QueryEstimateMultiHopSwapResponse
    }

    fn_query! {
        pub get_estimate_place_limit_order["/neutron.dex.Query/EstimatePlaceLimitOrder"]: QueryEstimatePlaceLimitOrderRequest => QueryEstimatePlaceLimitOrderResponse
    }

    fn_query! {
        pub get_pool["/neutron.dex.Query/Pool"]: QueryPoolRequest => QueryPoolResponse
    }

    fn_query! {
        pub get_pool_by_id["/neutron.dex.Query/PoolByID"]: QueryPoolByIdRequest => QueryPoolResponse
    }

    fn_query! {
        pub get_pool_metadata["/neutron.dex.Query/PoolMetadata"]: QueryGetPoolMetadataRequest => QueryGetPoolMetadataResponse
    }

    fn_query! {
        pub get_pool_metadata_all["/neutron.dex.Query/PoolMetadataAll"]: QueryAllPoolMetadataRequest => QueryAllPoolMetadataResponse
    }
}
