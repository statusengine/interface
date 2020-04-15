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

use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\ScheduleddowntimeServiceLoaderInterface;
use Statusengine\ValueObjects\ScheduleddowntimeQueryOptions;

class ScheduleddowntimeServiceLoader implements ScheduleddowntimeServiceLoaderInterface {

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * ScheduleddowntimeHostLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    public function getScheduledServicedowntimes(ScheduleddowntimeQueryOptions $QueryOptions) {
        $fields = [
            'hostname',
            'service_description',
            'author_name',
            'comment_data',
            'internal_downtime_id',
            'scheduled_start_time',
            'scheduled_end_time',
            'duration',
            'node_name',
            'was_started'
        ];

        $baseQuery = sprintf('SELECT %s FROM statusengine_service_scheduleddowntimes WHERE 1=1 ', implode(',', $fields));

        if ($QueryOptions->getHostnameLike() != '') {
            $baseQuery = sprintf('%s AND hostname ~* ?', $baseQuery);
        }

        if ($QueryOptions->getServicedescriptionLike() != '') {
            $baseQuery = sprintf('%s AND service_description ~* ?', $baseQuery);
        }

        $baseQuery .= $this->getClusterNameQuery($QueryOptions);

        $baseQuery = sprintf(
            '%s ORDER BY %s %s LIMIT ? OFFSET ?',
            $baseQuery,
            $QueryOptions->getOrder(),
            $QueryOptions->getDirection()
        );

        $query = $this->Backend->prepare($baseQuery);

        $i = 1;
        if ($QueryOptions->getHostnameLike() != '') {
            $like = sprintf('.*%s.*', $QueryOptions->getHostnameLike());
            $query->bindValue($i++, $like);
        }

        if ($QueryOptions->getServicedescriptionLike() != '') {
            $like = sprintf('.*%s.*', $QueryOptions->getServicedescriptionLike());
            $query->bindValue($i++, $like);
        }

        foreach ($QueryOptions->getClusterName() as $clusterName) {
            $query->bindValue($i++, $clusterName);
        }

        $query->bindValue($i++, $QueryOptions->getLimit(), \PDO::PARAM_INT);
        $query->bindValue($i++, $QueryOptions->getOffset(), \PDO::PARAM_INT);

        $results = $this->Backend->fetchAll($query);

        return $results;

    }

    /**
     * @param ScheduleddowntimeQueryOptions $QueryOptions
     * @return string
     */
    private function getClusterNameQuery(ScheduleddowntimeQueryOptions $QueryOptions) {
        $placeholders = [];
        foreach ($QueryOptions->getClusterName() as $clusterName) {
            $placeholders[] = '?';
        }
        if (!empty($placeholders)) {
            return sprintf(' AND node_name IN(%s)', implode(',', $placeholders));
        }
        return '';
    }

    /**
     * @param $hostDowntime
     * @return array
     */
    public function getScheduledServicedowntimesByHostdowntime($hostDowntime) {
        $baseQuery = 'SELECT * FROM statusengine_service_scheduleddowntimes WHERE hostname=? AND scheduled_start_time=? AND scheduled_end_time=? AND node_name=?';


        $query = $this->Backend->prepare($baseQuery);
        $i = 1;
        $query->bindParam($i++, $hostDowntime['hostname']);
        $query->bindParam($i++, (int)$hostDowntime['scheduled_start_time'], \PDO::PARAM_INT);
        $query->bindParam($i++, (int)$hostDowntime['scheduled_end_time'], \PDO::PARAM_INT);
        $query->bindParam($i++, $hostDowntime['node_name']);

        return $this->Backend->fetchAll($query);
    }
}
