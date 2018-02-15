<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Loader\Crate;

use Crate\PDO\PDO;
use Statusengine\Backend\Crate\Crate;
use Statusengine\Backend\StorageBackend;
use Statusengine\Controller\Host;
use Statusengine\Loader\HostLoaderInterface;
use Statusengine\ValueObjects\HostQueryOptions;
use Statusengine\ValueObjects\HostSearchQueryOptions;

class HostLoader implements HostLoaderInterface {

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * HostLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }


    public function getHostList(HostQueryOptions $HostQueryOptions) {
        $baseQuery = 'SELECT * FROM statusengine_hoststatus';
        $useStateFilter = false;
        if ($HostQueryOptions->sizeOfStateFilter() > 0 && $HostQueryOptions->sizeOfStateFilter() < 3) {
            $useStateFilter = true;
            $baseQuery = sprintf('%s WHERE current_state IN(%s)', $baseQuery, implode(',', $HostQueryOptions->getStateFilter()));
        }

        if ($HostQueryOptions->getHostnameLike() != '') {
            $sql = ($useStateFilter) ? 'AND' : 'WHERE';
            $useStateFilter = true;
            $baseQuery = sprintf(' %s %s hostname ~* ? ', $baseQuery, $sql);
        }

        $baseQuery .= $this->getClusterNameQuery($HostQueryOptions, !$useStateFilter);

        $baseQuery = sprintf(
            '%s ORDER BY %s %s LIMIT ? OFFSET ?',
            $baseQuery,
            $HostQueryOptions->getOrder(),
            $HostQueryOptions->getDirection()
        );

        $query = $this->Backend->prepare($baseQuery);

        $i = 1;
        if ($HostQueryOptions->getHostnameLike() != '') {
            $like = sprintf('.*%s.*', $HostQueryOptions->getHostnameLike());
            $query->bindValue($i++, $like);
        }

        foreach ($HostQueryOptions->getClusterName() as $clusterName) {
            $query->bindValue($i++, $clusterName);
        }

        $query->bindValue($i++, $HostQueryOptions->getLimit(), PDO::PARAM_INT);
        $query->bindValue($i++, $HostQueryOptions->getOffset(), PDO::PARAM_INT);
        $hostResult = $this->Backend->fetchAll($query);
        $hostNames = $this->extractField('hostname', $hostResult);
        $serviceOverview = $this->getServiceStateCount($hostNames);


        $result = [];
        foreach ($hostResult as $record) {
            if(isset($serviceOverview[$record['hostname']])){
                $record['servicestatus_overview'] = $serviceOverview[$record['hostname']];
            }else{
                $record['servicestatus_overview'] = ['ok' => 0, 'warning' => 0, 'critical' => 0, 'unknown' => 0];
            }
            $result[] = $record;
        }
        return $result;

    }

    /**
     * @param HostQueryOptions $HostQueryOptions
     * @return array
     */
    public function getHostDetails(HostQueryOptions $HostQueryOptions) {
        $fields = [
            'notifications_enabled',
            'active_checks_enabled',
            'passive_checks_enabled',
            'flap_detection_enabled',
            'event_handler_enabled',
            'is_flapping',
            'is_hardstate',
            'problem_has_been_acknowledged',
            'last_check',
            'next_check',
            'last_state_change',
            'hostname',
            'node_name',
            'current_check_attempt',
            'max_check_attempts',
            'current_state',
            'output',
            'long_output',
            'perfdata',
            'check_timeperiod',
            'normal_check_interval',
            'retry_check_interval',
            'scheduled_downtime_depth',
            'status_update_time'
        ];

        $query = $this->Backend->prepare(sprintf('SELECT %s FROM statusengine_hoststatus WHERE hostname=?', implode(',', $fields)));
        $query->bindValue(1, $HostQueryOptions->getHostname());

        $hostResult = $this->Backend->fetchAll($query);

        $result['hoststatus'] = [];
        if (isset($hostResult[0])) {
            $result['hoststatus'] = $hostResult[0];
        }
        $result['servicestatus'] = $this->getServicesByHost($HostQueryOptions);
        return $result;
    }

    /**
     * @param $hostNames
     * @return array
     */
    public function getServiceStateCount($hostNames) {
        if (empty($hostNames)) {
            return [];
        }

        $placeholders = [];
        for ($i = 1; $i <= sizeof($hostNames); $i++) {
            $placeholders[] = '?';
        }

        $query = $this->Backend->prepare(sprintf('SELECT hostname, current_state, COUNT(current_state) AS counter
        FROM statusengine_servicestatus WHERE hostname IN(%s) GROUP BY hostname, current_state', implode(',', $placeholders)));

        $i = 1;
        foreach ($hostNames as $hostName) {
            $query->bindValue($i++, $hostName);
        }

        $return = [];
        foreach ($this->Backend->fetchAll($query) as $record) {
            $return[$record['hostname']][$record['current_state']] = $record['counter'];
        }
        return $this->convertToFrontendServicestatus($return);
    }

    public function search(HostSearchQueryOptions $HostSearchQueryOptions){
        $query = 'SELECT hostname FROM statusengine_hoststatus WHERE hostname ~* ? ORDER BY hostname ASC LIMIT ?';
        $query = $this->Backend->prepare($query);

        $i = 1;
        $like = sprintf('.*%s.*', $HostSearchQueryOptions->getHostnameLike());
        $query->bindValue($i++, $like);
        $query->bindValue($i++, $HostSearchQueryOptions->getLimit());

        return $this->Backend->fetchAll($query);
    }

    /**
     * @param string $field
     * @param array $resultData
     * @return array
     */
    public function extractField($field, $resultData) {
        $return = [];
        foreach ($resultData as $record) {
            $return[] = $record[$field];
        }
        return $return;
    }

    /**
     * @param $servicestatus
     * @return array
     */
    private function convertToFrontendServicestatus($servicestatus) {
        $result = [];
        foreach ($servicestatus as $hostname => $record) {
            $stateCount = [];
            foreach ([0 => 'ok', 1 => 'warning', 2 => 'critical', 3 => 'unknown'] as $stateKey => $stateName) {
                $stateCount[$stateName] = 0;
                if (isset($record[$stateKey])) {
                    $stateCount[$stateName] = $record[$stateKey];
                }
            }
            $result[$hostname] = $stateCount;
        }
        return $result;
    }

    /**
     * @param HostQueryOptions $QueryOptions
     * @param bool $useWhere
     * @return string
     */
    private function getClusterNameQuery(HostQueryOptions $QueryOptions, $useWhere = true) {
        $operator = 'WHERE';
        if (!$useWhere) {
            $operator = 'AND';
        }
        $placeholders = [];
        foreach ($QueryOptions->getClusterName() as $clusterName) {
            $placeholders[] = '?';
        }
        if (!empty($placeholders)) {
            return sprintf(' %s node_name IN(%s)', $operator, implode(',', $placeholders));
        }
        return '';
    }


    /**
     * @param HostQueryOptions $HostQueryOptions
     * @return array
     */
    private function getServicesByHost(HostQueryOptions $HostQueryOptions) {
        $fields = [
            'notifications_enabled',
            'active_checks_enabled',
            'passive_checks_enabled',
            'flap_detection_enabled',
            'event_handler_enabled',
            'is_flapping',
            'is_hardstate',
            'last_check',
            'next_check',
            'last_state_change',
            'hostname',
            'service_description',
            'node_name',
            'current_check_attempt',
            'max_check_attempts',
            'current_state',
            'output',
            'long_output',
            'perfdata',
            'check_timeperiod',
            'normal_check_interval',
            'retry_check_interval',
            'scheduled_downtime_depth',
            'problem_has_been_acknowledged'
        ];

        $baseQuery = sprintf('SELECT %s FROM statusengine_servicestatus WHERE hostname=? ', implode(',', $fields));
        if ($HostQueryOptions->sizeOfServiceStateFilter() > 0 && $HostQueryOptions->sizeOfServiceStateFilter() < 4) {
            $baseQuery = sprintf('%s AND current_state IN(%s)', $baseQuery, implode(',', $HostQueryOptions->getServiceStateFilter()));
        }

        if ($HostQueryOptions->getServiceDescriptionLike() != '') {
            $baseQuery = sprintf('%s AND service_description ~* ? ', $baseQuery);
        }

        $baseQuery = sprintf('%s ORDER BY current_state DESC, service_description ASC', $baseQuery);

        $query = $this->Backend->prepare($baseQuery);
        $query->bindValue(1, $HostQueryOptions->getHostname());

        if ($HostQueryOptions->getServiceDescriptionLike() != '') {
            $query->bindValue(2, sprintf('.*%s.*', $HostQueryOptions->getServiceDescriptionLike()));
        }

        $query->bindValue(1, $HostQueryOptions->getHostname());
        return $this->Backend->fetchAll($query);
    }

}