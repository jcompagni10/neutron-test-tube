use cosmrs::proto::cosmos::bank::v1beta1::{
    MsgSend, MsgSendResponse, QueryAllBalancesRequest, QueryAllBalancesResponse,
    QueryBalanceRequest, QueryBalanceResponse, QueryTotalSupplyRequest, QueryTotalSupplyResponse,
};

use test_tube::module::Module;
use test_tube::runner::Runner;
use test_tube::{fn_execute, fn_query};

pub struct Bank<'a, R: Runner<'a>> {
    runner: &'a R,
}

impl<'a, R: Runner<'a>> Module<'a, R> for Bank<'a, R> {
    fn new(runner: &'a R) -> Self {
        Self { runner }
    }
}

impl<'a, R> Bank<'a, R>
where
    R: Runner<'a>,
{
    fn_execute! {
        pub send: MsgSend["/cosmos.bank.v1beta1.MsgSend"] => MsgSendResponse
    }

    fn_query! {
        pub query_balance ["/cosmos.bank.v1beta1.Query/Balance"]: QueryBalanceRequest => QueryBalanceResponse
    }

    fn_query! {
        pub query_all_balances ["/cosmos.bank.v1beta1.Query/AllBalances"]: QueryAllBalancesRequest => QueryAllBalancesResponse
    }

    fn_query! {
        pub query_total_supply ["/cosmos.bank.v1beta1.Query/TotalSupply"]: QueryTotalSupplyRequest => QueryTotalSupplyResponse
    }
}

#[cfg(test)]
mod test {
    use cosmwasm_std::{coin, coins, Coin, CosmosMsg};
    use test_tube::{utils::coins_to_proto, Account};

    use crate::NeutronTestApp;

    use super::*;

    #[test]
    fn bank_module_works() {
        let app = NeutronTestApp::default();
        let acc = app
            .init_accounts(
                &vec![
                    coin(500_000_000_000_000, "uusd"),
                    coin(500_000_000_000_000, "untrn"),
                ],
                2,
            )
            .unwrap();

        let bank = Bank::new(&app);

        let QueryAllBalancesResponse { balances, .. } = bank
            .query_all_balances(&QueryAllBalancesRequest {
                address: acc[0].address().clone(),
                pagination: None,
            })
            .unwrap();

        assert_eq!(
            balances,
            coins_to_proto(&[
                Coin::new(500_000_000_000_000, "untrn"),
                Coin::new(500_000_000_000_000, "uusd")
            ])
        );

        app.execute_cosmos_msgs::<MsgSendResponse>(
            &[CosmosMsg::Bank(cosmwasm_std::BankMsg::Send {
                to_address: acc[1].address(),
                amount: coins(400_000_000_000_000, "uusd"),
            })],
            &acc[0],
        )
        .unwrap();

        let QueryAllBalancesResponse { balances, .. } = bank
            .query_all_balances(&QueryAllBalancesRequest {
                address: acc[1].address(),
                pagination: None,
            })
            .unwrap();

        assert_eq!(
            balances,
            coins_to_proto(&[
                Coin::new(900_000_000_000_000, "uusd"),
                Coin::new(500_000_000_000_000, "untrn")
            ])
        );
    }
}
