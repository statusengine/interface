<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler

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
use Statusengine\Loader\LogentryLoaderInterface;
use Statusengine\ValueObjects\LogentryQueryOptions;
use Statusengine\ValueObjects\QueryOptions;

class LogentryLoader implements LogentryLoaderInterface {

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * LogentryLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }


    /**
     * @return array
     */
    public function getLogentries(LogentryQueryOptions $QueryOptions) {
        if ($QueryOptions->getEntryTimeGt()) {
            $query = $this->getLogentriesGtFromTimestamp($QueryOptions);
            return $this->Backend->fetchAll($query);
        }

        if ($QueryOptions->getEntryTimeLt()) {
            $query = $this->getLogentriesLtFromTimestamp($QueryOptions);
            return $this->Backend->fetchAll($query);
        }

        if ($QueryOptions->getLogentryDataLike() != '') {
            $baseQuery = 'SELECT * FROM statusengine_logentries WHERE logentry_data ~* ? %s ORDER BY entry_time DESC LIMIT ?';
            $baseQuery = sprintf($baseQuery, $this->getClusterNameQuery($QueryOptions, false));
            $query = $this->Backend->prepare($baseQuery);

            $i = 1;
            $like = sprintf('.*%s.*', $QueryOptions->getLogentryDataLike());
            $query->bindValue($i++, $like);

            foreach ($QueryOptions->getClusterName() as $clusterName) {
                $query->bindValue($i++, $clusterName);
            }

            $query->bindValue($i++, $QueryOptions->getLimit(), PDO::PARAM_INT);
            return $this->Backend->fetchAll($query);
        }

        $baseQuery = 'SELECT * FROM statusengine_logentries %s ORDER BY entry_time DESC LIMIT ?';
        $baseQuery = sprintf($baseQuery, $this->getClusterNameQuery($QueryOptions));

        $query = $this->Backend->prepare($baseQuery);
        $i = 1;
        foreach ($QueryOptions->getClusterName() as $clusterName) {
            $query->bindValue($i++, $clusterName);
        }

        $query->bindValue($i++, $QueryOptions->getLimit(), PDO::PARAM_INT);

        return $this->Backend->fetchAll($query);
    }

    /**
     * @param LogentryQueryOptions $QueryOptions
     * @return array|bool|\Crate\PDO\PDOStatement|\PDOStatement
     */
    private function getLogentriesGtFromTimestamp(LogentryQueryOptions $QueryOptions) {
        if ($QueryOptions->getLogentryDataLike() != '') {
            $baseQuery = 'SELECT * FROM statusengine_logentries WHERE logentry_data ~* ?  %s AND entry_time > ? ORDER BY entry_time DESC LIMIT ?';
            $baseQuery = sprintf($baseQuery, $this->getClusterNameQuery($QueryOptions, false));
            $query = $this->Backend->prepare($baseQuery);

            $i = 1;
            $like = sprintf('.*%s.*', $QueryOptions->getLogentryDataLike());
            $query->bindValue($i++, $like);

            foreach ($QueryOptions->getClusterName() as $clusterName) {
                $query->bindValue($i++, $clusterName);
            }

            $query->bindValue($i++, $QueryOptions->getEntryTimeGt(), PDO::PARAM_INT);
            $query->bindValue($i++, $QueryOptions->getLimit(), PDO::PARAM_INT);
            return $query;
        }

        $i = 1;
        $baseQuery = 'SELECT * FROM statusengine_logentries WHERE entry_time > ? %s ORDER BY entry_time DESC LIMIT ?';
        $baseQuery = sprintf($baseQuery, $this->getClusterNameQuery($QueryOptions, false));

        $query = $this->Backend->prepare($baseQuery);
        $query->bindValue($i++, $QueryOptions->getEntryTimeGt(), PDO::PARAM_INT);

        foreach ($QueryOptions->getClusterName() as $clusterName) {
            $query->bindValue($i++, $clusterName);
        }

        $query->bindValue($i++, $QueryOptions->getLimit(), PDO::PARAM_INT);
        return $query;
    }

    /**
     * @param LogentryQueryOptions $QueryOptions
     * @return bool|\Crate\PDO\PDOStatement|\PDOStatement
     */
    private function getLogentriesLtFromTimestamp(LogentryQueryOptions $QueryOptions) {
        if ($QueryOptions->getLogentryDataLike() != '') {
            $baseQuery = 'SELECT * FROM statusengine_logentries WHERE logentry_data ~* ? %s AND entry_time < ? ORDER BY entry_time DESC LIMIT ?';
            $baseQuery = sprintf($baseQuery, $this->getClusterNameQuery($QueryOptions, false));
            $i = 1;
            $query = $this->Backend->prepare($baseQuery);

            $like = sprintf('.*%s.*', $QueryOptions->getLogentryDataLike());
            $query->bindValue($i++, $like);

            foreach ($QueryOptions->getClusterName() as $clusterName) {
                $query->bindValue($i++, $clusterName);
            }

            $query->bindValue($i++, $QueryOptions->getEntryTimeLt(), PDO::PARAM_INT);
            $query->bindValue($i++, $QueryOptions->getLimit(), PDO::PARAM_INT);
            return $query;
        }

        $i = 1;
        $baseQuery = 'SELECT * FROM statusengine_logentries WHERE entry_time < ? %s ORDER BY entry_time DESC LIMIT ?';
        $baseQuery = sprintf($baseQuery, $this->getClusterNameQuery($QueryOptions, false));

        $query = $this->Backend->prepare($baseQuery);
        $query->bindValue($i++, $QueryOptions->getEntryTimeLt());

        foreach ($QueryOptions->getClusterName() as $clusterName) {
            $query->bindValue($i++, $clusterName);
        }

        $query->bindValue($i++, $QueryOptions->getLimit(), PDO::PARAM_INT);
        return $query;
    }

    /**
     * @param LogentryQueryOptions $QueryOptions
     * @param bool $useWhere
     * @return string
     */
    private function getClusterNameQuery(LogentryQueryOptions $QueryOptions, $useWhere = true) {
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

}